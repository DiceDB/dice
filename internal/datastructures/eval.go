package datastructures

// EvalResponse represents the response of an evaluation operation for a command from store.
// It contains the sequence ID, the result of the store operation, and any error encountered during the operation.
type EvalResponse struct {
	Result interface{} // Result holds the outcome of the Store operation. Currently, it is expected to be of type []byte, but this may change in the future.
	Error  error       // Error holds any error that occurred during the operation. If no error, it will be nil.
}

// Following functions should be used to create a new EvalResponse with the given result and error.
// These ensure that result and error are mutually exclusive.
// If result is nil, then error should be non-nil and vice versa.

// makeEvalResult creates a new EvalResponse with the given result and nil error.
// This is a helper function to create a new EvalResponse with the given result and nil error.
/**
 * @param {interface{}} result - The result of the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given result and nil error.
 */
func MakeEvalResult(result interface{}) *EvalResponse {
	return &EvalResponse{
		Result: result,
		Error:  nil,
	}
}

// makeEvalError creates a new EvalResponse with the given error and nil result.
// This is a helper function to create a new EvalResponse with the given error and nil result.
/**
 * @param {error} err - The error that occurred during the store operation.
 * @returns {EvalResponse} A new EvalResponse with the given error and nil result.
 */
func MakeEvalError(err error) *EvalResponse {
	return &EvalResponse{
		Result: nil,
		Error:  err,
	}
}
