package game

import "encoding/json"

type RemoveCommand struct {
	Point *Point
	BoardCommand
}

func (r *RemoveCommand) GetCommandType() CommandType {
	return REMOVE
}

func (r *RemoveCommand) MarshalJSON() ([]byte, error) {
	type k struct {
		X       int         `json:"x"`
		Y       int         `json:"y"`
		Command CommandType `json:"command"`
	}
	return json.Marshal(k{r.Point.X, r.Point.Y, r.GetCommandType()})
}
