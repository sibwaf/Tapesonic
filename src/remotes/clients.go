package remotes

import "tapesonic/subsonic"

func GetSubsonicAuthMethod(credentials *RemoteCredentials) subsonic.AuthMethod {
	if credentials == nil {
		return &subsonic.EmptyAuth{}
	} else {
		plainAuth := subsonic.NewPlainAuth(credentials.Username, credentials.Password)
		return &plainAuth
	}
}
