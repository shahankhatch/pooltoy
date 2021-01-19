package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/google/uuid"
)

const TypeMsgCreateUser = "msg_create_user"

var _ sdk.Msg = &MsgCreateUser{}

type MsgCreateUser1 struct {
	ID          string
	Creator     sdk.AccAddress `json:"creator" yaml:"creator"`
	UserAccount sdk.AccAddress `json:"userAccount" yaml:"userAccount"`
	IsAdmin     bool           `json:"isAdmin" yaml:"isAdmin"`
	Name        string         `json:"name" yaml:"name"`
	Email       string         `json:"email" yaml:"email"`
}

func NewMsgCreateUser(creator sdk.AccAddress, userAccount sdk.AccAddress, isAdmin bool, name string, email string) MsgCreateUser {
	return MsgCreateUser{
		Id:          uuid.New().String(),
		Creator:     creator.String(),
		UserAccount: userAccount.String(),
		IsAdmin:     isAdmin,
		Name:        name,
		Email:       email,
	}
}

func (msg *MsgCreateUser) CreatorAccAddress() sdk.AccAddress {
	accCreator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic("Invalid address found")
	}
	return accCreator
}

func (msg *MsgCreateUser) UserAddressAccAddress() sdk.AccAddress {
	accUserAccount, err := sdk.AccAddressFromBech32(msg.UserAccount)
	if err != nil {
		panic("Invalid address found")
	}
	return accUserAccount
}

func (msg *MsgCreateUser) Route() string {
	return RouterKey
}

func (msg *MsgCreateUser) Type() string {
	return TypeMsgCreateUser
}

func (msg *MsgCreateUser) GetSigners() []sdk.AccAddress {
	acc, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{acc}
}

func (msg *MsgCreateUser) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateUser) ValidateBasic() error {
	accCreator, _ := sdk.AccAddressFromBech32(msg.Creator)
	if accCreator.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "creator can't be empty")
	}
	accUser, _ := sdk.AccAddressFromBech32(msg.UserAccount)
	if accUser.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "UserAccount can't be empty")
	}
	if msg.Name == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Name can't be empty")
	}
	return nil
}
