# Design Spec
link: https://drive.google.com/file/d/1gI_dwf2h8irzAGefHfuX_79ASWItdqch/view?usp=sharing

## Key
key is 32 bytes length with:
- 10 bytes prefix
- 20 bytes key

## Value
value is depend on type of state object

## State Object
- Assume that hash(something) return 32 bytes value
1. All Shard Committee
- key: first 12 bytes of `hash(shard-committee-prefix)` with first 20 bytes of `hash(beaconheight)`
- value: 
    * temporary value: 32 bytes random ID
    * real value: all shard committee
- key -> temporary value -> real value. Beacause shard committee not change so frequently