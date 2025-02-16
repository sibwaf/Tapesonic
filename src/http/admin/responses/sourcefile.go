package responses

import "tapesonic/storage"

type SourceFileRs struct {
	Codec string
}

func SourceFileToSourceFileRs(file storage.SourceFile) SourceFileRs {
	return SourceFileRs{
		Codec: file.Codec,
	}
}
