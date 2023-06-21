package hasher

import (
	"fmt"
	"testing"
)

var testValue = []struct {
	in          string
	expected256 string
	expected512 string
}{
	{
		"secret",
		"2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
		"bd2b1aaf7ef4f09be9f52ce2d8d599674d81aa9d6a4421696dc4d93dd0619d682ce56b4d64a9ef097761ced99e0f67265b5f76085e5b0ee7ca4696b2ad6fe2b2",
	},
	{
		"anothersecret",
		"1b34d471266588be57fa753a4dd2494463d5cb137139b637a71ad5644240aa15",
		"0736ae735df2aa577ddb482705915cd6cc084df395ae19f9647bd7f39f77be2c1fec63a2bda1528e0891a2663681a18925a6b1988ce1c4a35eb20e83c418646f",
	},
}

func TestHashPassword(t *testing.T) {
	for _, val := range testValue {
		actual256, err := HashPassword(val.in)
		if err != nil || actual256 != val.expected256 {
			t.Errorf("HashPassword(\"%s\") = \"%s\", expected \"%s\", error %v", val.in, actual256, val.expected256, err)
		}
	}
}

func TestHashEmptyPassword(t *testing.T) {
	if res, err := HashPassword(""); err != ErrEmptyPassword {
		t.Errorf("HashPassword(\"\") = \"%s\", expected \"\", error %v, expected %v", res, err, ErrEmptyPassword)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	if res := CheckPasswordHash(testValue[0].in, testValue[0].expected256); !res {
		t.Errorf("CheckPasswordHash(%s, %s) = %v, expected %v", testValue[0].in, testValue[0].expected256, res, true)
	}

	if res := CheckPasswordHash(testValue[0].in, testValue[1].expected256); res {
		t.Errorf("CheckPasswordHash(%s, %s) = %v, expected %v", testValue[0].in, testValue[1].expected256, res, false)
	}
}

func TestHasherInterface(t *testing.T) {
	t.Run("Unsupported hash type", func(t *testing.T) {
		_, err := New(3)
		if err != ErrUnsupportedHashType {
			t.Errorf("New(3), expected error %v, got %v", ErrUnsupportedHashType, err)
		}
	})

	hi, err := New(HashSHA512)
	if err != nil {
		t.Errorf("New(HashSHA512) = %v, expected error %v", hi, err)
	}

	t.Run("HashPassword HashSHA512", func(t *testing.T) {
		for _, val := range testValue {
			actual512, err := hi.HashPassword(val.in)
			if err != nil || actual512 != val.expected512 {
				t.Errorf("Hasher interface HashPassword(\"%s\") = \"%s\", expected \"%s\", error %v", val.in, actual512, val.expected512, err)
			}
		}
	})

	t.Run("CheckPasswordHash HashSHA512", func(t *testing.T) {
		if res := hi.CheckPasswordHash(testValue[0].in, testValue[0].expected512); !res {
			t.Errorf("CheckPasswordHash(%s, %s) = %v, expected %v", testValue[0].in, testValue[0].expected512, res, true)
		}

		if res := hi.CheckPasswordHash(testValue[0].in, testValue[1].expected512); res {
			t.Errorf("CheckPasswordHash(%s, %s) = %v, expected %v", testValue[0].in, testValue[1].expected512, res, false)
		}
	})
}

func ExampleHashPassword() {
	pwd := "secret"
	hash, _ := HashPassword(pwd)
	fmt.Println(pwd)
	fmt.Println(hash)
	// Output:
	// secret
	// 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
}

func ExampleCheckPasswordHash() {
	pwd := "secret"
	fmt.Println(CheckPasswordHash(pwd, "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b"))
	fmt.Println(CheckPasswordHash(pwd, "1b34d471266588be57fa753a4dd2494463d5cb137139b637a71ad5644240aa15"))
	// Output:
	// true
	// false
}

func ExampleHasher() {
	hi, _ := New(HashSHA512)
	pwd := "secret"
	hash, _ := hi.HashPassword(pwd)
	fmt.Println(pwd)
	fmt.Println(hash)
	fmt.Println(hi.CheckPasswordHash(pwd, "bd2b1aaf7ef4f09be9f52ce2d8d599674d81aa9d6a4421696dc4d93dd0619d682ce56b4d64a9ef097761ced99e0f67265b5f76085e5b0ee7ca4696b2ad6fe2b2"))
	fmt.Println(hi.CheckPasswordHash(pwd, "0736ae735df2aa577ddb482705915cd6cc084df395ae19f9647bd7f39f77be2c1fec63a2bda1528e0891a2663681a18925a6b1988ce1c4a35eb20e83c418646f"))
	// Output:
	// secret
	// bd2b1aaf7ef4f09be9f52ce2d8d599674d81aa9d6a4421696dc4d93dd0619d682ce56b4d64a9ef097761ced99e0f67265b5f76085e5b0ee7ca4696b2ad6fe2b2
	// true
	// false
}
