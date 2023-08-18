package device

type Command int64

const (
	SWITCH_HOST Command = iota
	VIDEO_WALL
	GET_HOST
	GET_INFO
	GET_SIGNAL
)

func getCommandString(cmdType Command) string {
	switch cmdType {
	case SWITCH_HOST:
		return "CMD=START&UNIT.ID=ALL&STREAM.HOST=temp&VW.ACTIVE=FALSE&STREAM.CONNECT=TRUE&CMD=END"
	case VIDEO_WALL:
		return "CMD=START&UNIT.ID=ALL&VW.MAX_ROWS=temp&VW.MAX_COLUMNS=temp&VW.ROW=temp&VW.COLUMN=temp&VW.ACTIVE=TRUE&CMD=END"
	case GET_HOST:
		return "CMD=START&UNIT.ID=ALL&QUERY.KEY=STREAM.HOST&CMD=END"
	case GET_INFO:
		return "CMD=START&UNIT.ID=ALL&QUERY.KEY=UNIT.MODEL&QUERY.KEY=UNIT.FIRMWARE&QUERY.KEY=UNIT.FIRMWARE_DATE&QUERY.KEY=IP.ADDRESS&QUERY.KEY=UNIT.MAC_ADDRESS&CMD=END"
	case GET_SIGNAL:
		return "CMD=START&UNIT.ID=ALL&QUERY.VIDEO_TIMING=TRUE&CMD=END"
	default:
		return ""
	}
}

type videoWallParams struct {
	ColumnPosition int `json:"columnPosition"`
	RowPosition    int `json:"rowPosition"`
	TotalColumns   int `json:"totalColumns"`
	TotalRows      int `json:"totalRows"`
}
