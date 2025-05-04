package responses

import (
	"testing"
)

/*
This test may seem simple, it is an appropriate level of testing for a constructor function.
The key purpose of such a test is to ensure that the struct is being initialized correctly
with the given inputs. This is especially important to verify as part of a regression suite,
ensuring that future changes do not inadvertently break the struct initialization.
*/
func TestNewLoginResponse(t *testing.T) {
	token := "access-token"
	refreshToken := "refresh-token"
	exp := 3600

	lr := NewLoginResponse(token, refreshToken, exp)

	if lr.AccessToken != token {
		t.Errorf("expected AccessToken to be %s, got %s", token, lr.AccessToken)
	}
	if lr.RefreshToken != refreshToken {
		t.Errorf("expected RefreshToken to be %s, got %s", refreshToken, lr.RefreshToken)
	}
	if lr.Exp != exp {
		t.Errorf("expected Exp to be %d, got %d", exp, lr.Exp)
	}
}
