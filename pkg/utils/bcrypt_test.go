package utils

import (
	"strings"
	"testing"
)

func TestHashPassword_Normal(t *testing.T) {
	hash, err := HashPassword("mySecurePass123")
	if err != nil {
		t.Fatalf("正常密码应成功: %v", err)
	}
	if hash == "" {
		t.Fatal("哈希值不应为空")
	}
	if hash == "mySecurePass123" {
		t.Fatal("哈希值不应等于明文")
	}
}

func TestHashPassword_Empty(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("空密码也应能哈希: %v", err)
	}
	if hash == "" {
		t.Fatal("空密码的哈希值不应为空")
	}
}

func TestHashPassword_Long(t *testing.T) {
	longPwd := strings.Repeat("a", 72) // bcrypt 最大 72 字节
	hash, err := HashPassword(longPwd)
	if err != nil {
		t.Fatalf("72 字节密码应成功: %v", err)
	}
	if hash == "" {
		t.Fatal("长密码的哈希值不应为空")
	}
}

func TestHashPassword_TooLong(t *testing.T) {
	longPwd := strings.Repeat("a", 73) // 超过 bcrypt 限制
	_, err := HashPassword(longPwd)
	if err != nil {
		t.Logf("超过 72 字节返回错误（预期行为）: %v", err)
		return
	}
	t.Log("超过 72 字节未报错（取决于 bcrypt 实现）")
}

func TestCheckPassword_Correct(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("哈希失败: %v", err)
	}
	if !CheckPassword("correct", hash) {
		t.Fatal("正确密码应校验通过")
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("哈希失败: %v", err)
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("错误密码不应校验通过")
	}
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	if CheckPassword("anything", "") {
		t.Fatal("空哈希值应校验失败")
	}
}

func TestRoundTrip(t *testing.T) {
	passwords := []string{"hello", "P@ssw0rd!", "测试中文密码", "a b c 1 2 3"}
	for _, pwd := range passwords {
		hash, err := HashPassword(pwd)
		if err != nil {
			t.Fatalf("密码 %q 哈希失败: %v", pwd, err)
		}
		if !CheckPassword(pwd, hash) {
			t.Fatalf("密码 %q 往返校验失败", pwd)
		}
	}
}
