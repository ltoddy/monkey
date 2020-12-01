package visitor

type Config struct {
	FollowRedirect bool
	HttpMethod     string
	Include        bool
	MaxRedirects   int
}
