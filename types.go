package agentphone

import(
	"net/url"
	"strconv"
)

// ListParams holds common pagination parameters accepted by most "list"
// endpoints (List Agents, List Numbers, List Calls, ...).

// NOTE: AgentPhone's exact pagination convention (limit/offset vs.
// page/page_size vs. cursor-based) hasn't been confirmed against a live
// response yet. This assumes the common limit/offset convention; adjust
// the field names and toQuery() below once a real response is checked
// against docs.agentphone.ai.

