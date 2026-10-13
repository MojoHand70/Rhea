package adapter

import (
	"context"

	"rhea/internal/core"
	"rhea/internal/store"
)

// The registry, read as two lists (DIRECTION 2026-10-10): what this company
// has connected — every declared channel, joined with the code the binary
// carries — and what Rhea can read — every adapter the binary carries,
// joined with the channel that binds it. The shell and the CLI render the
// same status; neither knows a particular channel.

// ChannelStatus is one declared channel and where it stands.
type ChannelStatus struct {
	Name          string              `json:"name"`
	Version       int                 `json:"version"`
	Status        string              `json:"status"`
	Description   string              `json:"description"`
	Adapter       string              `json:"adapter"`
	CodeVersion   int                 `json:"code_version"`
	Format        string              `json:"format"`
	Identity      string              `json:"identity"`
	Confidence    core.ConfidenceSpec `json:"confidence"`
	Events        []string            `json:"events"`
	Volume        string              `json:"volume"`
	Market        string              `json:"market,omitempty"`
	Implemented   int                 `json:"implemented"` // code version the binary carries; 0 when none
	State         string              `json:"state"`
	EventsEmitted int                 `json:"events_emitted"`
}

// AdapterStatus is one adapter the binary carries and whether a channel
// connects it.
type AdapterStatus struct {
	Implemented
	Channel        string `json:"channel,omitempty"`
	ChannelVersion int    `json:"channel_version,omitempty"`
	State          string `json:"state"`
}

// RegistryStatus is both lists.
type RegistryStatus struct {
	Channels []ChannelStatus `json:"channels"`
	Adapters []AdapterStatus `json:"adapters"`
}

// The states a channel or an adapter can be in.
const (
	StateConnected       = "connected"
	StateNotImplemented  = "not implemented"
	StateVersionMismatch = "version mismatch"
	StateNotConnected    = "not connected"
	StateAwaiting        = "awaiting approval"
)

// Registry joins the declared channels with the implemented adapters. Active
// channels come first, then the latest versions that are not the active one
// (drafts at the gate, retired channels), the way the verb list is read.
func Registry(ctx context.Context, s *store.Store) (RegistryStatus, error) {
	active, err := s.ActiveActivities(ctx)
	if err != nil {
		return RegistryStatus{}, err
	}
	latest, err := s.LatestActivities(ctx)
	if err != nil {
		return RegistryStatus{}, err
	}
	counts, err := s.ActivityEventCounts(ctx)
	if err != nil {
		return RegistryStatus{}, err
	}
	var out RegistryStatus
	activeVersion := map[string]int{}
	connected := map[string]ChannelStatus{} // by adapter name: the active channel binding it
	add := func(a core.Activity) {
		ch := a.Spec.Channel
		st := ChannelStatus{
			Name: a.Name, Version: a.Version, Status: a.Status, Description: a.Description,
			Adapter: ch.Adapter, CodeVersion: ch.CodeVersion, Format: ch.Format, Identity: ch.Identity,
			Confidence: ch.Confidence, Events: ch.Events, Volume: ch.Volume, Market: ch.Market, EventsEmitted: counts[a.Name],
		}
		if len(st.Events) == 0 {
			st.Events = []string{a.Spec.Emits}
		}
		if st.Volume == "" {
			st.Volume = core.VolumeDocuments
		}
		impl, ok := implemented[ch.Adapter]
		if ok {
			st.Implemented = impl.CodeVersion
		}
		switch {
		case a.Status != core.StatusActive:
			st.State = StateAwaiting
			if a.Status != core.StatusDraft {
				st.State = a.Status
			}
		case !ok:
			st.State = StateNotImplemented
		case impl.CodeVersion != ch.CodeVersion:
			st.State = StateVersionMismatch
		default:
			st.State = StateConnected
		}
		if a.Status == core.StatusActive {
			connected[ch.Adapter] = st
		}
		out.Channels = append(out.Channels, st)
	}
	for _, a := range active {
		activeVersion[a.Name] = a.Version
		if a.Spec.Channel != nil {
			add(a)
		}
	}
	for _, a := range latest {
		if a.Spec.Channel != nil && a.Version != activeVersion[a.Name] {
			add(a)
		}
	}
	for _, impl := range ImplementedAdapters() {
		st := AdapterStatus{Implemented: impl, State: StateNotConnected}
		if ch, ok := connected[impl.Name]; ok {
			st.Channel, st.ChannelVersion = ch.Name, ch.Version
			st.State = ch.State
		}
		out.Adapters = append(out.Adapters, st)
	}
	if out.Channels == nil {
		out.Channels = []ChannelStatus{}
	}
	return out, nil
}
