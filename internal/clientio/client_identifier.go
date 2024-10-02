package clientio

type ClientIdentifier struct {
	ClientIdentifierID int
	IsHTTPClient       bool
}

func NewClientIdentifier(clientIdentifierID int, isHTTPClient bool) ClientIdentifier {
	return ClientIdentifier{
		ClientIdentifierID: clientIdentifierID,
		IsHTTPClient:       isHTTPClient,
	}
}
