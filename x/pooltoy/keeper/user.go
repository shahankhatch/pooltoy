package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

func (k Keeper) CreateUser(ctx sdk.Context, user types.User) {

	store := ctx.KVStore(k.storeKey)
	key := []byte(types.UserPrefix + user.ID)
	value := k.Cdc.LegacyAmino.MustMarshalBinaryLengthPrefixed(user)
	store.Set(key, value)

	acc := k.accountKeeper.GetAccount(ctx, user.UserAccount)
	if acc == nil {
		acc = k.accountKeeper.NewAccountWithAddress(ctx, user.UserAccount)
		k.accountKeeper.SetAccount(ctx, acc)
	}

}

func (k Keeper) GetUserByAccAddress(ctx sdk.Context, queriedUserAccAddress sdk.AccAddress) types.User {
	store := ctx.KVStore(k.storeKey)

	var queriedUser types.User

	iterator := sdk.KVStorePrefixIterator(store, []byte(types.UserPrefix))
	for ; iterator.Valid(); iterator.Next() {
		var user types.User
		k.Cdc.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(store.Get(iterator.Key()), &user)
		if user.UserAccount.Equals(queriedUserAccAddress) {
			queriedUser = user
		}
	}
	return queriedUser
}

func (k Keeper) ListUsers(ctx sdk.Context) ([]byte, error) {
	var userList []types.User
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.UserPrefix))
	for ; iterator.Valid(); iterator.Next() {
		var user types.User
		k.Cdc.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(store.Get(iterator.Key()), &user)
		userList = append(userList, user)
	}
	res := codec.MustMarshalJSONIndent(k.Cdc.LegacyAmino, userList)
	return res, nil
}
