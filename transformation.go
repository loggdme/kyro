package kyro

func NilIfDefault[T comparable](v T) *T {
	if v == *new(T) {
		return nil
	}
	return &v
}

func NilIfPointerDefault[T comparable](v *T) *T {
	if v == nil || *v == *new(T) {
		return nil
	}
	return v
}

func NullishCoalescing[T comparable](v *T, defaultValue *T) *T {
	if v == nil || *v == *new(T) {
		return defaultValue
	}
	return v
}

type WeightedProportionCheck struct {
	Score     int
	Condition bool
}

func CalculateWeightedProportion(checks []WeightedProportionCheck) float64 {
	maxScore, currentScore := 0, 0
	for _, check := range checks {
		maxScore += check.Score
		if check.Condition {
			currentScore += check.Score
		}
	}

	normalizedScore := 0.0
	if maxScore > 0 {
		normalizedScore = float64(currentScore) / float64(maxScore)
	}

	return normalizedScore
}

type WeightedSumCheck struct {
	Weight float64
	Value  float64
}

func CalculateWeightedSum(checks []WeightedSumCheck) float64 {
	weightedSum := 0.0

	for _, check := range checks {
		weightedSum += check.Weight * check.Value
	}

	return weightedSum
}
