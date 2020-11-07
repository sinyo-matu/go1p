package go1p

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignIn(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "")
	assert.Nil(t, err)
}

func TestGetItemFullly(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "")
	assert.Nil(t, err)
	result, err := cli.GetItemFully("Amazon")
	fmt.Println(result.Details.Fields[0].Designation, result.Details.Fields[0].Value)
	assert.Nil(t, err)
}

func TestGetItemLitely(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "")
	assert.Nil(t, err)
	result, err := cli.GetUsernameAndPassword("SBI証券取引パス")
	fmt.Println(result.Password)
	assert.Nil(t, err)
}

func TestGetSessionLastTime(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "")
	assert.Nil(t, err)
	result, err := cli.GetSessionLastTime()
	fmt.Println(result)
	assert.Nil(t, err)
}

func TestGetItemListByCategories(t *testing.T) {
	cli := NewCli()
	err := cli.SignInWithPresetPass("my", "")
	assert.Nil(t, err)
	result, err := cli.GetItemListByCategories("Identity")
	fmt.Println(result)
	assert.Nil(t, err)
}