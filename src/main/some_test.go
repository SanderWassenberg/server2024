package main


import (
    "testing"
    "src/pwhash"
    "strings"
    "encoding/json"
)

func Test_EmailValidating(t *testing.T) {
	var values = []struct {
		email string
		validity bool
	}{
		{"asd@asd.asd", true},
		{"asd@asd", false},
		{"@asd.asd", false},
		{".asd@asd.asd", false},
	}

	for _, v := range values {
		got_validity := is_valid_email(v.email)
		fail := got_validity != v.validity
		if fail { t.Error(v.email, "should have validity", v.validity, "got", got_validity) }
	}
}

func Test_PasswordHashing(t *testing.T) {

	pwhash.Iterations = 1
	pwhash.Threads    = 1
	pwhash.Memory_KiB = 8
	pwhash.KeyLen     = 16
	err := pwhash.ValidateSettings()
	if err != nil {
		t.Fatal("invalid argon2 settings for this test:", err)
	}

	password := "hiya!"

	hash, _ := pwhash.EncodePassword(password)
	match := pwhash.VerifyPassword(password, hash)

	if !match {
		t.Fatal("password validation/generation failed")
	}
}

func Test_PasswordHashInterpreting(t *testing.T) {
	pwhash.Iterations = 2
	pwhash.Threads    = 4
	pwhash.Memory_KiB = 16
	pwhash.KeyLen     = 8

	// hash with different settings than the ones above
	hash := "$argon2id$v=19$m=8,t=1,p=1$CUlmIKLzXsIlVPe/Dv+0/g$3dcKAuUnALa6h6NWdgzjvQ"

	// if verification fails, probably different settings where used than the ones listed here.
	match := pwhash.VerifyPassword("hiya!", hash)

	if !match {
		t.Fatal("password validation/generation failed")
	}
}

func Test_UTF8LengthValidation_ExactLength(t *testing.T) {
	text := "Y񔧖m豱֐㎀輗̿瘶˰趻؇񗎑ɒz񽙁󓑟󢅍ãჺv󒪶Oё󶈣Ϣ"
	expected_len := 26
	limit := 26

	len, ok := verify_utf8_string_len(text, limit) // length should be allowed, just not OVER 33
	if !ok {
		t.Fatal("length should be ok but was not, got length", len, "with limit", limit)
	}

	if len != expected_len {
		t.Fatal("unexpected length:", len)
	}
}

func Test_UTF8LengthValidation_TooLong(t *testing.T) {
	text := "Y񔧖m豱֐㎀輗̿瘶˰趻؇񗎑ɒz񽙁󓑟󢅍ãჺv󒪶Oё󶈣Ϣ"
	expected_len := 26
	limit := 25

	len, ok := verify_utf8_string_len(text, limit) // length should be allowed, just not OVER 33
	if ok {
		t.Fatal("length should be too long but was deemed ok, got length", len, "with limit", limit)
	}

	if len != expected_len {
		t.Fatal("unexpected length:", len)
	}
}

func Test_UserInfoNotSensitive(t *testing.T) {
	info := UserInfo{}
	info_json, err := json.Marshal(info)
	if err != nil { t.Fatal("json marshal failed") }

	fields := make(map[string]any)
	json.Unmarshal(info_json, &fields)

	forbidden_terms := []string{"id", "session", "token", "banned", "admin", "secret"}

	for key, _ := range fields {
		key = strings.ToLower(key)
		for _, term := range forbidden_terms {
			if strings.Contains(key, term) {
				t.Error("json marshalling userinfo exposed field with sensitive term:", term, "\nfull marshal:\n", string(info_json))
			}
		}
	}
}

func Test_ValidateUsername(t *testing.T) {
	values := []struct {
		username string
		valid bool
	}{
		{"_asfsdf_dfsdf_",                true},
		{"kek😃",                          false},
		{"<script>alert('xss')</script>", false},
		{"no spaces",                     false},
		{"OR 1 = 1 --",                   false},
		{"OR 1 = 1;",                     false},
		{"OR 1 = 1",                      false},
		{"JustAName26",                   true},
		{"_xXx_Killer_xXx_",              true},
		{"",                              false},
	}

	for _, val := range values {
		got_valid := is_valid_username(val.username)
		success := got_valid == val.valid
		if !success {
			t.Error("username", val.username, "should have validity", val.valid)
		}
	}
}