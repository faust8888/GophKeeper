package common

import pb "github.com/faust8888/GophKeeper/internal/pkg/proto/gophkeeperpb"

func SecretTypeFromString(s string) pb.SecretType {
	switch s {
	case "EMPTY":
		return pb.SecretType_EMPTY
	case "UNSPECIFIED":
		return pb.SecretType_UNSPECIFIED
	case "CREDENTIALS":
		return pb.SecretType_CREDENTIALS
	case "TEXT":
		return pb.SecretType_TEXT
	case "BINARY":
		return pb.SecretType_BINARY
	case "CARD":
		return pb.SecretType_CARD
	default:
		return pb.SecretType_UNSPECIFIED
	}
}

func SecretTypeToString(secretType pb.SecretType) string {
	switch secretType {
	case pb.SecretType_EMPTY:
		return "EMPTY"
	case pb.SecretType_UNSPECIFIED:
		return "UNSPECIFIED"
	case pb.SecretType_CREDENTIALS:
		return "CREDENTIALS"
	case pb.SecretType_TEXT:
		return "TEXT"
	case pb.SecretType_BINARY:
		return "BINARY"
	case pb.SecretType_CARD:
		return "CARD"
	default:
		return "UNSPECIFIED"
	}
}
