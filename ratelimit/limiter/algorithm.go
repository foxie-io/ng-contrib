package limiter

/*
Rate limiting algorithms comparison:

	{
		| Algorithm          | Burst | Smooth | Complexity      | Memory          | DoS / Abuse Resistance           |
		| ------------------ | ----- | ------ | --------------- | --------------- | ------------------------------- |
		| Fixed Window       | ❌     | ❌      | O(1)            | O(1)            | ❌ prone to spikes at window boundary |
		| Sliding Window     | ❌     | ✅      | O(n per client) | O(n per client) | ⚠️ high memory per client; can be exploited by many clients (high cardinality) |
		| Token Bucket       | ✅     | ✅      | O(1)            | O(1)            | ✅ handles bursts well; constant memory; better for abuse traffic |

	}

Notes:

- "Burst": whether the algorithm allows short-term bursts.

- "Smooth": whether rate limiting is evenly spread or can spike.

- "Complexity": per-request time complexity.

- "Memory": per-client memory usage.

- "DoS / Abuse Resistance": safety under high request rate or high client cardinality.

Use Token Bucket for high-throughput, public-facing APIs; Sliding Window for accurate, low-volume scenarios.
*/
type Algorithm string

const (
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmFixedWindow   Algorithm = "fixed_window"
	AlgorithmSlidingWindow Algorithm = "sliding_window"
)

func (a Algorithm) String() string {
	return string(a)
}
