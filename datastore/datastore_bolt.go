package datastore

import (
	"bytes"
	"strconv"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	rm "github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/utils/time"
	bolt "go.etcd.io/bbolt"
)

// routeCollection will contain the routing info for a DB
var routeCollection []byte = []byte("routeCollection")

// scheduleCollection will contain all the schedules and will be used to fetch the minute wise jobs
var scheduleCollection []byte = []byte("scheduleCollection")

// It uses BoltDB which uses B+tree implementation.
// The data is stored in the below format
//	- routeCollection (contains routes for this DB)
//	- scheduleCollection (contains minute wise buckets for all the collections)
// 		- minutewise buckets
// 			- timestamp : uniqueJobID
//	- user job collection 1
//	- user job collection 2
//	- user job collection n
type BoltDataStore struct {
	db *bolt.DB

	// The Database name
	dbName string

	// Data dbPath
	dbPath string
}

func CreateBoltDataStore(dbName string, path string) (*BoltDataStore, error) {
	if path == "" {
		path = "bolt-db/"
	}
	path += dbName + ".db"
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	return &BoltDataStore{
		db:     db,
		dbName: dbName,
		dbPath: path,
	}, err
}

func (bds *BoltDataStore) Close() {
	bds.db.Close()
}

func (bds *BoltDataStore) GetJob(collection, jobID string) (*jm.Job, error) {
	// Start the transaction.
	tx, err := bds.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the job from the collection
	bkt, err := tx.CreateBucketIfNotExists([]byte(collection))
	if err != nil {
		return nil, err
	}

	val := bkt.Get([]byte(jobID))
	if val == nil {
		return nil, ErrKeyNotFound
	}

	return jm.GetJobFromBytes(val)
}

func (bds *BoltDataStore) SetJob(collection string, job *jm.Job) error {
	by, err := job.ToBytes()
	if err != nil {
		return err
	}

	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch the collection bucket
	bkt, err := tx.CreateBucketIfNotExists([]byte(collection))
	if err != nil {
		return err
	}

	// Insert the job in collection bucket
	err = bkt.Put([]byte(job.ID), by)
	if err != nil {
		return err
	}

	// Fetch the schedule collection bucket
	scheduleBkt, err := tx.CreateBucketIfNotExists(
		scheduleCollection)
	if err != nil {
		return err
	}

	// Fetch the minutewise bucket in scheduleBkt
	minuteBkt, err := scheduleBkt.CreateBucketIfNotExists(
		job.GetMinuteBucketName())
	if err != nil {
		return err
	}

	// Insert the schedule in the minute wise bucket
	uniqueKey := job.GetUniqueKey(collection)
	if err = minuteBkt.Put(uniqueKey, []byte("1")); err != nil {
		return err
	}

	if err = scheduleBkt.Put([]byte(job.ID), by); err != nil {
		return err
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}

func (bds *BoltDataStore) DeleteJob(collection, jobID string) error {
	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte(collection))
	if err != nil {
		return err
	}
	val := bkt.Get([]byte(jobID))
	if val == nil {
		return ErrKeyNotFound
	}
	job, _ := jm.GetJobFromBytes(val)

	// Fetch the schedule collection bucket
	scheduleBkt, err := tx.CreateBucketIfNotExists(
		scheduleCollection)
	if err != nil {
		return err
	}

	// Fetch the minutewise bucket in scheduleBkt
	minuteBkt, err := scheduleBkt.CreateBucketIfNotExists(
		job.GetMinuteBucketName())
	if err != nil {
		return err
	}

	// Delete the schedule in the minute wise bucket
	uniqueKey := job.GetUniqueKey(collection)
	if err = minuteBkt.Delete(uniqueKey); err != nil {
		return err
	}

	// Delete the job from collection
	if err = bkt.Delete([]byte(jobID)); err != nil {
		return err
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}

// FetchJobTill is used to fetch all the jobs in the datastore till the provided time
func (bds *BoltDataStore) FetchJobTill(collection string, timeTill int) ([]*jm.Job, error) {
	// Start the transaction.
	tx, err := bds.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	c := tx.Bucket(scheduleCollection).Cursor()

	// TODO: Consider proper job key
	min := []byte(strconv.Itoa(time.GetCurrentMillis()))
	max := []byte(strconv.Itoa(timeTill))

	var jobs []*jm.Job

	// Iterate from min to max
	for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
		// Read the byte values to job struct
		j, err := jm.GetJobFromBytes(v)
		if err != nil {
			return nil, err
		}

		// Add all the jobs to list
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (bds *BoltDataStore) GetRoute(routeID string) (*rm.Route, error) {
	// Start the transaction.
	tx, err := bds.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists(routeCollection)
	if err != nil {
		return nil, err
	}

	val := bkt.Get([]byte(routeID))
	if val == nil {
		return nil, ErrKeyNotFound
	}

	return rm.GetRouteFromBytes(val)
}
func (bds *BoltDataStore) SetRoute(route *rm.Route) error {
	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists(routeCollection)
	if err != nil {
		return err
	}

	by, err := route.ToBytes()
	if err != nil {
		return err
	}

	// Insert the route byte array
	err = bkt.Put([]byte(route.ID), by)
	if err != nil {
		return err
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}

func (bds *BoltDataStore) DeleteRoute(route string) error {
	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists(routeCollection)
	if err != nil {
		return err
	}

	// Delete the route byte array
	err = bkt.Delete([]byte(route))
	if err != nil {
		return err
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}
