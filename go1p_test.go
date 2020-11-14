package go1p

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignIn(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "", true)
	fmt.Println(cli.expirationTime)
	assert.Nil(t, err)
}

func TestGetItemFullly(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "", true)
	assert.Nil(t, err)
	result, err := cli.GetItemFully("Amazon")
	fmt.Println(result.Details.Fields[0].Designation, result.Details.Fields[0].Value)
	assert.Nil(t, err)
}

func TestGetItemLitely(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "", true)
	assert.Nil(t, err)
	result, err := cli.GetUsernameAndPassword("SBI証券取引パス")
	fmt.Println(result.Password)
	assert.Nil(t, err)
}

func TestGetItemWithCustomizedFiled(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "", true)
	assert.Nil(t, err)
	result, err := cli.GetItemWithCustomizedField("SBIBank", "username")
	fmt.Println(string(result))
	assert.Nil(t, err)
}

func TestGetListWithFlag(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "", true)
	assert.Nil(t, err)
	result, err := cli.GetListWithFlag([]string{"categories"}, []string{"Secure Note,Bank Account"})
	assert.Nil(t, err)
	fmt.Println(result)
}
