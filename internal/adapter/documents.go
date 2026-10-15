package adapter

// What every reader of the document store shares (DIRECTION 2026-10-10,
// "Documents in"): the type a received file is, and the two events a
// reading ends in. Market readers (markets/<cc>) and the generic ones
// (MT940, the model-read PDF) speak these; the kernel defines none of them.
const (
	// ReceivedDocumentType is the documents pack's conventional vocabulary
	// (packs/common/documents.json) the readers scan — a type the pack
	// ships, like task_schedule for the clock; the kernel does not know it.
	ReceivedDocumentType = "received_document"

	// EventDocumentRead says a channel turned a received document into
	// typed events; the documents pack moves the document to "read" on it.
	EventDocumentRead = "document.read"
	// EventDocumentUnreadable says a channel could not: the reason is for a
	// person, so the event waits in the worklist.
	EventDocumentUnreadable = "document.unreadable"
)
