package worker

/** ====================================================================================
 * 🏁 Types & Constants
 * =====================================================================================
 */

type DeleteSessionsPayload struct {
	JTIs []string `json:"jtis"`
}

const TaskDeleteSessions = "cache:delete_sessions"
