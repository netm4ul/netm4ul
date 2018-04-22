package api

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
)

//CodeToResult defines a simple way to get predefined and coherents results from an error code.
var CodeToResult map[Code]Result

func init() {
	CodeToResult = make(map[Code]Result)

	CodeToResult = map[Code]Result{
		CodeOK:                 Result{Code: CodeOK, Status: "success"},
		CodeNotFound:           Result{Code: CodeNotFound, Status: "error", Message: "Ressource not found on the server !"},
		CodeCouldNotDecodeJSON: Result{Code: CodeCouldNotDecodeJSON, Status: "error", Message: "Could not decode provided json : invalid json"},
		CodeDatabaseError:      Result{Code: CodeDatabaseError, Status: "error", Message: "Something unexepected has happened with the database"},
		CodeNotImplementedYet:  Result{Code: CodeNotImplementedYet, Status: "error", Message: "Not implemented yet"},
	}
}
