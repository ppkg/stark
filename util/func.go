package util

func RunFunc(fn func() error, retryCount ...int) error {
	count := 3
	if len(retryCount) > 0 && retryCount[0] > 0 {
		count = retryCount[0]
	}
	var err error
	for i := 0; i < count; i++ {
		err = fn()
		if err == nil {
			break
		}
	}
	return err
}