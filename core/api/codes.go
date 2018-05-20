package api

import (
	"encoding/json"
	"net/http"
)

type Code int

const (
	//CodeOK : Successful operation
	CodeOK Code = iota
	//CodeNotFound : Ressource not found on the server
	CodeNotFound
	//CodeCouldNotDecodeJSON : Invalid json
	CodeCouldNotDecodeJSON
	//CodeDatabaseError : Something unexepected has happened with the database
	CodeDatabaseError
	//CodeNotImplementedYet : This function is not implemented yet.
	CodeNotImplementedYet
	//CodeInvalidInput : Unexepected inputs where provided
	CodeInvalidInput
	//CodeServerError : A server error occured
	CodeServerError
	//CodeForbidden : Access to this ressource forbidden
	CodeForbidden
)

//CodeToResult defines a simple way to get predefined and coherents results from an error code.
var CodeToResult map[Code]Result

func init() {
	CodeToResult = make(map[Code]Result)

	CodeToResult = map[Code]Result{
		CodeOK:                 {Code: CodeOK, Status: "success", HTTPCode: http.StatusOK},
		CodeNotFound:           {Code: CodeNotFound, Status: "error", Message: "Ressource not found on the server !", HTTPCode: http.StatusNotFound},
		CodeCouldNotDecodeJSON: {Code: CodeCouldNotDecodeJSON, Status: "error", Message: "Could not decode provided json : invalid json", HTTPCode: http.StatusBadRequest},
		CodeDatabaseError:      {Code: CodeDatabaseError, Status: "error", Message: "Something unexepected has happened with the database", HTTPCode: http.StatusInternalServerError},
		CodeNotImplementedYet:  {Code: CodeNotImplementedYet, Status: "error", Message: "Not implemented yet", HTTPCode: http.StatusNotImplemented},
		CodeInvalidInput:       {Code: CodeInvalidInput, Status: "error", Message: "Unexpected inputs", HTTPCode: http.StatusUnprocessableEntity}, //not perfect http status
		CodeServerError:        {Code: CodeServerError, Status: "error", Message: "A server error occured", HTTPCode: http.StatusInternalServerError},
		CodeForbidden:          {Code: CodeForbidden, Status: "error", Message: "Access forbidden", HTTPCode: http.StatusForbidden},
	}
}

func sendDefaultValue(w http.ResponseWriter, code Code) {
	res := CodeToResult[code]
	w.WriteHeader(CodeToResult[code].HTTPCode)
	json.NewEncoder(w).Encode(res)
}
