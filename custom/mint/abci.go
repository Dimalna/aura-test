package mint

import (
	"errors"
	"time"

	custommint "github.com/aura-nw/aura/custom/mint/keeper"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k custommint.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// check over max supply
	maxSupplyString := k.GetMaxSupply(ctx)
	maxSupply, ok := sdk.NewIntFromString(maxSupplyString)
	if !ok {
		panic(errors.New("panic convert max supply string to bigInt"))
	}
	k.Logger(ctx).Info("Get max supply from aura", "maxSupply", maxSupply.String())
	currentSupply := k.GetSupply(ctx, params.MintDenom)
	k.Logger(ctx).Info("Get current supply from network", "currentSupply", currentSupply.String())

	if currentSupply.LT(maxSupply) {
		// recalculate inflation rate
		totalStakingSupply := k.StakingTokenSupply(ctx)
		bondedRatio := k.BondedRatio(ctx)
		minter.Inflation = minter.NextInflationRate(params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
		k.SetMinter(ctx, minter)

		// mint coins, update supply
		mintedCoin := minter.BlockProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		supplyNext := currentSupply.Add(mintedCoin.Amount)
		if supplyNext.GT(maxSupply) {
			mintedCoin.Amount = maxSupply.Sub(currentSupply)
			mintedCoins = sdk.NewCoins(mintedCoin)
		}
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
				sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
				sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)

	} else {
		k.Logger(ctx).Info("Over the max supply", "currentSupply", currentSupply)
	}
}
