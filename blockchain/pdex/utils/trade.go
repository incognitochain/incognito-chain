package utils

import (
    "errors"
    "math/big"
)

func Quote(amount0 uint64, reserve0 uint64, reserve1 uint64) (uint64, error) {
    if amount0 <= 0 {
        return 0, errors.New("insufficient amount")
    }
    if reserve0 <= 0 || reserve1 <= 0 {
        return 0, errors.New("insufficient liquidity")
    }
    amount := big.NewInt(0).SetUint64(amount0)
    result := big.NewInt(0).Mul(amount, big.NewInt(0).SetUint64(reserve1))
    result.Div(result, big.NewInt(0).SetUint64(reserve0))
    if !result.IsUint64() {
        return 0, errors.New("number out of range uint64")
    }
    return result.Uint64(), nil
}

func CalculateBuyAmount(amountIn uint64, reserveIn uint64, reserveOut uint64, virtualReserveIn uint64, virtualReserveOut uint64, fee uint64) (uint64, error) {
    if amountIn <= 0 {
        return 0, errors.New("insufficient input amount")
    }
    if reserveIn <= 0 || reserveOut <= 0 {
        return 0, errors.New("insufficient liquidity")
    }
    amount := big.NewInt(0).SetUint64(amountIn)
    amount.Sub(amount, big.NewInt(0).SetUint64(fee))
    num := big.NewInt(0).Mul(amount, big.NewInt(0).SetUint64(virtualReserveOut))
    den := big.NewInt(0).Add(amount, big.NewInt(0).SetUint64(virtualReserveIn))
    result := num.Div(num, den)
    if !result.IsUint64() {
        return 0, errors.New("number out of range uint64")
    }
    return result.Uint64(), nil
}

func CalculateAmountToSell(amountOut uint64, reserveIn uint64, reserveOut uint64, virtualReserveIn uint64, virtualReserveOut uint64, fee uint64) (uint64, error) {
    if amountOut <= 0 {
        return 0, errors.New("insufficient input amount")
    }
    if reserveIn <= 0 || reserveOut <= 0 {
        return 0, errors.New("insufficient liquidity")
    }
    num := big.NewInt(0).SetUint64(virtualReserveIn)
    num.Mul(num, big.NewInt(0).SetUint64(amountOut))
    den := big.NewInt(0).SetUint64(virtualReserveOut)
    den.Sub(den, big.NewInt(0).SetUint64(amountOut))
    result := num.Div(num, den)
    result.Add(result, big.NewInt(0).SetUint64(fee + 1))
    if !result.IsUint64() {
        return 0, errors.New("number out of range uint64")
    }
    return result.Uint64(), nil
}
