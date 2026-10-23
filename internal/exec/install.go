package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"rhea/internal/core"
)

// Installing a pack in a company (DIRECTION 2026-10-24, step 2): the core's
// doors install_pack and switch_off_pack append the facts, the core's rules
// keep them as the company's installed_pack records, and the kernel checks
// each against the catalog before it is appended — a pack is installed only
// where it is approved, installed per company, and what it needs is there;
// never twice at once; never switched off under a pack that needs it.

// install is one installed_pack record of the company acted for.
type install struct {
	id, pack, from, to, status string
}

// installs reads the company's installed_pack records (its scope alone).
func (x *Executor) installs(ctx context.Context) ([]install, error) {
	objs, err := x.Store.ObjectsByType(ctx, "installed_pack")
	if err != nil {
		return nil, err
	}
	out := make([]install, 0, len(objs))
	for _, o := range objs {
		s := func(k string) string { v, _ := o.State[k].(string); return v }
		out = append(out, install{id: o.ID, pack: s("pack"), from: s("from"), to: s("to"), status: s("status")})
	}
	return out, nil
}

// standing is the company's install of pack still in force, if any.
func standing(is []install, pack string) (install, bool) {
	for _, i := range is {
		if i.pack == pack && i.status == "installed" {
			return i, true
		}
	}
	return install{}, false
}

func (x *Executor) checkInstall(ctx context.Context, ev core.Event) error {
	var p struct {
		Pack string `json:"pack"`
		From string `json:"from"`
	}
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return err
	}
	if ev.Company == "" {
		return fmt.Errorf("a pack is installed in a company — act for one")
	}
	b, ok, err := x.Store.ApprovedPack(ctx, p.Pack)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no version of the pack %s is approved", p.Pack)
	}
	if b.Pack == nil || b.Pack.Install != core.InstallCompany {
		return fmt.Errorf("the pack %s runs for every company of the installation — there is nothing to install", p.Pack)
	}
	is, err := x.installs(ctx)
	if err != nil {
		return err
	}
	if i, ok := standing(is, p.Pack); ok {
		return fmt.Errorf("the pack %s is installed here from %s already", p.Pack, i.from)
	}
	for _, need := range b.Pack.Needs {
		nb, ok, err := x.Store.ApprovedPack(ctx, need)
		if err != nil {
			return err
		}
		if ok && (nb.Pack == nil || nb.Pack.Install == core.InstallInstallation) {
			continue // every company has it
		}
		if i, ok := standing(is, need); !ok || i.from > p.From {
			return fmt.Errorf("the pack %s needs the pack %s installed here from %s", p.Pack, need, p.From)
		}
	}
	return nil
}

func (x *Executor) checkSwitchOff(ctx context.Context, ev core.Event) error {
	var p struct {
		Installed string `json:"installed"`
		To        string `json:"to"`
	}
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return err
	}
	o, err := x.Store.GetObject(ctx, p.Installed)
	if err != nil {
		return err
	}
	pack, _ := o.State["pack"].(string)
	from, _ := o.State["from"].(string)
	if o.Type != "installed_pack" || o.State["status"] != "installed" {
		return fmt.Errorf("%s is not a pack installed here", p.Installed)
	}
	if p.To < from {
		return fmt.Errorf("the pack %s is installed from %s — it cannot stop before", pack, from)
	}
	is, err := x.installs(ctx)
	if err != nil {
		return err
	}
	for _, i := range is {
		if i.status != "installed" || i.pack == pack {
			continue
		}
		b, ok, err := x.Store.ApprovedPack(ctx, i.pack)
		if err != nil {
			return err
		}
		if ok && b.Pack != nil && slices.Contains(b.Pack.Needs, pack) {
			return fmt.Errorf("the pack %s needs %s — switch %s off first", i.pack, pack, i.pack)
		}
	}
	return nil
}
