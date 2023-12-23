package datastore

import (
	"bytes"
	"log"
	"strconv"

	jm "github.com/aarthikrao/timeMachine/models/jobmodels"
	bolt "go.etcd.io/bbolt"

	"github.com/aarthikrao/timeMachine/components/jobstore"
	jStore "github.com/aarthikrao/timeMachine/components/jobstore"
)

// Compile time validation for jobstore interface
var _ jStore.JobStoreConn = &boltDataStore{}

// scheduleCollection will contain all the schedules and will be used to fetch the minute wise jobs
var scheduleCollection []byte = []byte("scheduleCollection")

// It uses BoltDB which uses B+tree implementation.
// The data is stored in the below format
//   ∟ routeCollection (contains routes for this DB)
//   ∟ scheduleCollection (contains minute wise buckets for all the collections)
//       ∟ minutewise buckets
//          ∟ timestamp : uniqueJobID
//   ∟ user job collection 1
//   ∟ user job collection 2
//   ∟ user job collection n

type boltDataStore struct {
	db *bolt.DB

	dbFilePath string
}

func CreateBoltDataStore(path string) (jStore.JobFetcher, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Opened db instance at:", path)

	return &boltDataStore{
		db:         db,
		dbFilePath: path,
	}, err
}

func (bds *boltDataStore) Close() error {
	return bds.db.Close()
}

func (bds *boltDataStore) GetJob(collection, jobID string) (*jm.Job, error) {
	// Start the transaction.
	tx, err := bds.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the job from the collection
	bkt := tx.Bucket([]byte(collection))
	if bkt == nil {
		return nil, ErrBucketNotFound
	}

	val := bkt.Get([]byte(jobID))
	if val == nil {
		return nil, ErrKeyNotFound
	}

	return jm.GetJobFromBytes(val)
}

func (bds *boltDataStore) SetJob(collection string, job *jm.Job) error {
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

	// Add the job in collection bucket
	{
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
	}

	// Add the job in schedule bucket
	{
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
		if err = minuteBkt.Put(
			job.GetUniqueKey(collection),
			job.StringifyTriggerTime(),
		); err != nil {
			return err
		}
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}

func (bds *boltDataStore) DeleteJob(collection, jobID string) error {
	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch the job from collection bucket, get the job value
	// and delete entry from collection bucket
	var jobByteValue []byte
	{
		bkt, err := tx.CreateBucketIfNotExists([]byte(collection))
		if err != nil {
			return err
		}
		jobByteValue = bkt.Get([]byte(jobID))
		if jobByteValue == nil {
			return ErrKeyNotFound
		}

		// Delete the job from collection
		if err = bkt.Delete([]byte(jobID)); err != nil {
			return err
		}
	}

	// Parse the job from bytes
	job, err := jm.GetJobFromBytes(jobByteValue)
	if err != nil {
		return err
	}

	// Delete the job from schedule bucket.
	{
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
	}

	// Commit the transaction and check for error.
	return tx.Commit()
}

func (bds *boltDataStore) Type() jobstore.JobStoreType {
	return jobstore.Database
}

// FetchJobForBucket is used to fetch all the jobs in the datastore till the provided time
func (bds *boltDataStore) FetchJobForBucket(minute int) ([]*jm.Job, error) {
	// Start the transaction.
	tx, err := bds.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Fetch the schedule collection bucket
	scheduleBkt, err := tx.CreateBucketIfNotExists(
		scheduleCollection)
	if err != nil {
		return nil, err
	}

	// Fetch the minute bucket
	minuteBucket := scheduleBkt.Bucket([]byte(strconv.Itoa(minute)))
	if minuteBucket == nil {
		// It means there are no jobs for this minute
		return nil, nil
	}

	var jobs []*jm.Job

	// Fetch all the jobs in the bucket
	c := minuteBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		// TODO: Read the byte values to job struct
		jobDetails := bytes.Split(k, []byte("_"))
		if len(jobDetails) != 2 {
			return nil, ErrInvalidDataformat
		}
		collection := jobDetails[0]
		jobID := jobDetails[1]

		// Fetch the collection
		collectionBkt := tx.Bucket(collection)
		if collectionBkt == nil {
			continue
		}

		// Fetch the job
		val := collectionBkt.Get(jobID)
		j, err := jm.GetJobFromBytes(val)
		if err != nil {
			// TODO : Check return
			return nil, err
		}

		// Add all the jobs to list
		jobs = append(jobs, j)
	}

	return jobs, nil
}
