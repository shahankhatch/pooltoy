package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUser = "msg_user"

var _ sdk.Msg = &User{}

type User1 struct {
	Creator     sdk.AccAddress `json:"creator" yaml:"creator"`
	ID          string         `json:"id" yaml:"id"`
	UserAccount sdk.AccAddress `json:"userAccount" yaml:"userAccount"`
	IsAdmin     bool           `json:"isAdmin" yaml:"isAdmin"`
	Name        string         `json:"name" yaml:"name"`
	Email       string         `json:"email" yaml:"email"`
}

func (u *User) CreatorAccAddress() sdk.AccAddress {
	accCreator, err := sdk.AccAddressFromBech32(u.Creator)
	if err != nil {
		panic("Invalid address found")
	}
	return accCreator
}

func (u *User) UserAddressAccAddress() sdk.AccAddress {
	accUserAccount, err := sdk.AccAddressFromBech32(u.UserAccount)
	if err != nil {
		panic("Invalid address found")
	}
	return accUserAccount
}

//
//func (u *User) Reset() {
//	*u = User{}
//}
//
//func (u *User) String() string {
//	return proto.CompactTextString(u)
//}
//
//func (u *User) ProtoMessage() {}

func (*User) Route() string {
	return RouterKey
}

func (*User) Type() string {
	return TypeMsgUser
}

func (*User) ValidateBasic() error {
	return nil
}

func (u *User) GetSignBytes() []byte {
	return sdk.MustSortJSON(Cdc.MustMarshalJSON(&u))
}

func (u *User) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{u.UserAddressAccAddress()}
}
