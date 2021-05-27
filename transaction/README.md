# Notes

This is the note for developing transaction package. The devs should read these notes when implement new type of tx or new version of tx.

Based on the design of TxBase, TxBase can be treated as metadata.Transaction and can call the right version itself. The way it figures out which version to delegate into is simply checking its .Version attribute. 

However, if the txVersion has not implemented the required function, the chain may self-loop recursively and crash.

Some noted functions that need to be implemented when embed TxBase:
   - CheckAuthorizedSender
   - Init
   - Verify
   - ValidateTxSalary (must code this even if that version dont have initTxSalary)
   
When adding new transaction version, please take a look at NewTxPrivacyFromVersionNumber.