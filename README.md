# Runestone Golang library

Bitcoin Rune protocol golang implement library.

## How to use
 
```go
go get github.com/bxelab/runestone@latest
```
or
```go
git clone https://github.com/bxelab/runestone.git
cd runestone/
go mod tidy
cd cmd/runestonecli
go run .
```
### Etching

Define a new Rune named *STUDYZY.GMAIL.COM* 

```go
func testEtching() {
	runeName := "STUDYZY.GMAIL.COM"
	symbol := '曾'
	myRune, err := runestone.SpacedRuneFromString(runeName)
	if err != nil {
		fmt.Println(err)
		return
	}
	amt := uint128.From64(666666)
	ca := uint128.From64(21000000)
	etching := &runestone.Etching{
		Rune:    &myRune.Rune,
		Spacers: &myRune.Spacers,
		Symbol:  &symbol,
		Terms: &runestone.Terms{
			Amount: &amt,
			Cap:    &ca,
		},
	}
	r := runestone.Runestone{Etching: etching}
	data, err := r.Encipher()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Etching data: 0x%x\n", data)
	dataString, _ := txscript.DisasmString(data)
	fmt.Printf("Etching Script: %s\n", dataString)
}
```

### Mint

```go
func testMint() {
	runeIdStr := "2609649:946"
	runeId, _ := runestone.RuneIdFromString(runeIdStr)
	r := runestone.Runestone{Mint: runeId}
	data, err := r.Encipher()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Mint Rune[%s] data: 0x%x\n", runeIdStr, data)
	dataString, _ := txscript.DisasmString(data)
	fmt.Printf("Mint Script: %s\n", dataString)
}
```

### Decode

```go
func testDecode() {
	data, _ := hex.DecodeString("140114001600") //Mint UNCOMMON•GOODS
	var tx wire.MsgTx
	builder := txscript.NewScriptBuilder()
	// Push opcode OP_RETURN
	builder.AddOp(txscript.OP_RETURN)
	// Push MAGIC_NUMBER
	builder.AddOp(runestone.MAGIC_NUMBER)
	// Push payload
	builder.AddData(data)
	pkScript, _ := builder.Script()
	txOut := wire.NewTxOut(0, pkScript)
	tx.AddTxOut(txOut)
	r := &runestone.Runestone{}
	artifact, err := r.Decipher(&tx)
	if err != nil {
		fmt.Println(err)
		return
	}
	a, _ := json.Marshal(artifact)
	fmt.Printf("Artifact: %s\n", string(a))
}
```

### Reference:

* https://docs.ordinals.com/runes/specification.html
* https://github.com/ordinals/ord/
