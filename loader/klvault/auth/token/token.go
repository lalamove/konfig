package token

import "time"

// Token auth provider
type Token struct {
	T string
}

// Token returns a vault token or an error if it encountered one.
// {"jwt": "'"$KUBE_TOKEN"'", "role": "{{ SERVICE_ACCOUNT_NAME }}"}
func (k *Token) Token() (string, time.Duration, error) {
	return k.T, 10 * time.Second, nil
}
