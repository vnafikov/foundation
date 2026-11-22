package uploadingfileparameters

type LeaseExpiredBatch struct {
	TTLInterval   string
	LeaseInterval string
	BatchSize     int
	LeaseToken    string
}
