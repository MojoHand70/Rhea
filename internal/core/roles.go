package core

// The roles are the system's own, a closed list (KK, 2026-10-27: "Roles are
// more suitable for commercial ownerships of contracts and technical
// parameters setup … The rest we can manage by Groups"). A membership puts a
// person in a company with one of them; everything a business decides about
// who does what — who keeps the books, who approves a bank account — is a
// group's. Their rights are written here, not in data: a rule of a
// company's own could amend a record, and a role every company shares must
// not be one company's to widen.
const (
	RoleRheaDesigner   = "rhea_designer"   // designs Apps and runs the interviews
	RoleTenantOwner    = "tenant_owner"    // owns the tenant's contract with Rhea
	RoleOfficeManager  = "office_manager"  // runs an accounting office: its parameters and master data
	RoleCompanyManager = "company_manager" // owns a company, an office's client or not
	RoleUser           = "user"            // works in Rhea; what they do is their groups'
	RoleGuest          = "guest"           // from outside; what they do is their groups'
	RoleRheaSupport    = "rhea_support"    // Rhea's own staff; in a company's books by its timed grant alone
)

// Roles is every role, in the order a person reads them.
var Roles = []string{RoleRheaDesigner, RoleTenantOwner, RoleOfficeManager, RoleCompanyManager, RoleUser, RoleGuest, RoleRheaSupport}

// groupDoors are the doors that say who is in which group: a role's, never a
// group's (KK, 2026-10-27: the Rhea Designer, the Tenant Owner, the Office
// Manager and the Company Manager add a person to a group).
const groupDoors = "door:declare_group door:join_group door:leave_group"

// RoleRights is what each role gives in the company its membership names.
var RoleRights = map[string]string{
	// sees the company to design for it, releases its Apps, keeps its
	// groups; books nothing (Claude, 2026-10-27)
	RoleRheaDesigner: "view:* type:* door:approve_bundle door:reject_bundle " + groupDoors,
	// who is in, and nothing of the books by this role alone (KK, 2026-10-27)
	RoleTenantOwner:    "domain:people",
	RoleOfficeManager:  "*",
	RoleCompanyManager: "*",
	RoleUser:           "",
	RoleGuest:          "",
	// the books, never who may do what — the support role of 2026-10-21
	RoleRheaSupport: "door:* view:* type:*",
}

// IsRole reports whether a name is one of the roles.
func IsRole(name string) bool {
	_, ok := RoleRights[name]
	return ok
}

// A FormerRole is how a role of 2026-10-19, which grants already in the log
// name, reads now: the role it became and, for three, the group of the same
// name it holds as well (KK, 2026-10-27: "Chief Accountant, Accountant -
// should be groups"). The log is never rewritten; its old words are read.
type FormerRole struct{ Role, Group string }

// FormerRoles maps the five roles the people pack declared until v4.
var FormerRoles = map[string]FormerRole{
	"owner":            {RoleCompanyManager, ""},
	"support":          {RoleRheaSupport, ""},
	"chief_accountant": {RoleUser, "chief_accountant"},
	"accountant":       {RoleUser, "accountant"},
	"reader":           {RoleUser, "reader"},
}

// RoleOf reads the role a membership names: one of the roles as it is, a
// former role as what it became; ok is false for any other name, which
// grants nothing.
func RoleOf(name string) (role, group string, ok bool) {
	if IsRole(name) {
		return name, "", true
	}
	if f, former := FormerRoles[name]; former {
		return f.Role, f.Group, true
	}
	return "", "", false
}

// GroupAllUsers is the group Rhea keeps: every person a membership lets into
// the company, a guest apart. Nobody puts anyone into it; an App or another
// group names it to reach everyone (KK, 2026-10-27).
const GroupAllUsers = "all_users"
