#pragma once

#ifdef __cplusplus
extern "C" {
#endif
	void hpmt_basic_test(void* hpmtInfoPtr);
	void hpmt_basic_from_raw_test_1(void* hpmtInfoPtr);
	void hpmt_basic_from_raw_test_2(void* hpmtInfoPtr);
	void hpmt_5m_1m_batch_insertion_with_random_proof_test(void* hpmtInfoPtr);
	void hpmt_50m_1m_batch_insertion_test(void* hpmtPtr);
	void leveldb_performance_test(void* hpmtInfoPtr, char* dbFile);

#ifdef __cplusplus
}
#endif