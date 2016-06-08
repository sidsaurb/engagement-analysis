package conf

const (
	DbUser = "veea"
	DbPassword = "aeev"
	DbName = "veea"

	AdminUsername = "admin"

	// time after which a session expires (in seconds)
	SessionExpireTime int64 = 30 * 24 * 60 * 60
	// length of the session id
	SessionIdLength = 128

	// time after which a view expires (in seconds)
	ViewExpireTime int64 = 5 * 60 * 60
	// length of the view id
	ViewIdLength = 64

	BasePath = "/home/garvit/cs/go/work/src/github.com/gpahal/veea/"
)
