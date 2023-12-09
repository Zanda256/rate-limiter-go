package tokenbucket

type TokenBucketConfig struct {
	Cap    int
	Period int
}

type checkLimitInput struct {
	Ip     string
	UserID string
}

var store = map[string]int{}

func (config *TokenBucketConfig) checkBucket(in string) {

}
