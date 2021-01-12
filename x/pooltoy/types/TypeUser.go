package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &User{}

type User struct {
	Creator     sdk.AccAddress `json:"creator" yaml:"creator"`
	ID          string         `json:"id" yaml:"id"`
	UserAccount sdk.AccAddress `json:"userAccount" yaml:"userAccount"`
	IsAdmin     bool           `json:"isAdmin" yaml:"isAdmin"`
	Name        string         `json:"name" yaml:"name"`
	Email       string         `json:"email" yaml:"email"`
}

func (u User) Reset() {
	panic("implement me")
}

func (u User) String() string {
	panic("implement me")
}

func (u User) ProtoMessage() {
	panic("implement me")
}

func (u User) Route() string {
	panic("implement me")
}

func (u User) Type() string {
	panic("implement me")
}

func (u User) ValidateBasic() error {
	panic("implement me")
}

func (u User) GetSignBytes() []byte {
	panic("implement me")
}

func (u User) GetSigners() []sdk.AccAddress {
	panic("implement me")
}
