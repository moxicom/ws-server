package auth

var it uint64 = 0

func validateToken(token string) (uint64, error) {
	it += 1
	return it - 1, nil
}
