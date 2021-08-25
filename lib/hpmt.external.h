#pragma once

/*Adding headers for cgo*/
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

	//Start the trie 
	void* Start(char* configFile);

	// Latest CONFIRMED batch ID
	uint64_t GetLatestBatchID(void* hpmtPtr);

	// Insert hashes into the tire, must be 32 chars / 64 bytes long
	uint64_t BatchInsert(void* hpmtPtr, char* bytes, uint64_t count, char* rootHash);

	// Insert into the tire from raw data, hashes will be calculated internally
	uint64_t BatchInsertFromRaw(void* hpmtPtr, char* bytes, uint64_t* lenArray, uint64_t count, char* rootHash);

	bool Confirm(void* hpmtPtr, uint64_t batchId);

	// If the batch is still in cache, if no caller has to call RecomputeProofs() with original data instead
	bool IsInCache(void* hpmtPtr, uint64_t batchId); 

	// Get proofs from local cache 
	void GetProofs(void* hpmtPtr, uint64_t batchId, char* leafHash, char* proofs, uint64_t* count);

	// Recompute proofs, from original dataset provided
	void RecomputeProofs(void* hpmtPtr, char* bytes, uint64_t inputCount, uint64_t batchId, char* leafHash, char* proofs, uint64_t* proofCount); 

	// Get the log messages
	void GetLogMsg(void* hpmtInfoPtr, char* buffer, uint64_t bufferSize);

	// Stop the trie
	void Stop(void* hpmtPtr);

	// Get version 
	void GetVersion(char* version);

	// Get product name
	void GetProduct(char* product);

#ifdef __cplusplus
}
#endif