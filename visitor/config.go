package visitor

type Config struct {
	FollowRedirect bool
	Method         string
	Include        bool
	MaxRedirects   int
	Headers        []string
}
