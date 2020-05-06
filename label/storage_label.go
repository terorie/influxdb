package label

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/kv"
)

func (s *Store) CreateLabel(ctx context.Context, tx kv.Tx, l *influxdb.Label) error {
	v, err := json.Marshal(l)
	if err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	encodedID, err := l.ID.Encode()
	if err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	idx, err := tx.Bucket(labelIndex)
	if err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	key, err := labelIndexKey(l)
	if err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	if err := idx.Put([]byte(key), encodedID); err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	b, err := tx.Bucket(labelBucket)
	if err != nil {
		return err
	}

	if err := b.Put(encodedID, v); err != nil {
		return &influxdb.Error{
			Err: err,
		}
	}

	return nil
}

func labelMappingKey(m *influxdb.LabelMapping) ([]byte, error) {
	lid, err := m.LabelID.Encode()
	if err != nil {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Err:  err,
		}
	}

	rid, err := m.ResourceID.Encode()
	if err != nil {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Err:  err,
		}
	}

	key := make([]byte, influxdb.IDLength+influxdb.IDLength) // len(rid) + len(lid)
	copy(key, rid)
	copy(key[len(rid):], lid)

	return key, nil
}

// labelAlreadyExistsError is used when creating a new label with
// a name that has already been used. Label names must be unique.
func labelAlreadyExistsError(lbl *influxdb.Label) error {
	return &influxdb.Error{
		Code: influxdb.EConflict,
		Msg:  fmt.Sprintf("label with name %s already exists", lbl.Name),
	}
}

func labelIndexKey(l *influxdb.Label) ([]byte, error) {
	orgID, err := l.OrgID.Encode()
	if err != nil {
		return nil, &influxdb.Error{
			Code: influxdb.EInvalid,
			Err:  err,
		}
	}

	k := make([]byte, influxdb.IDLength+len(l.Name))
	copy(k, orgID)
	copy(k[influxdb.IDLength:], []byte(strings.ToLower((l.Name))))
	return k, nil
}

func (s *Service) unique(ctx context.Context, tx kv.Tx, indexBucket, indexKey []byte) error {
	bucket, err := tx.Bucket(indexBucket)
	if err != nil {
		return kv.UnexpectedIndexError(err)
	}

	_, err = bucket.Get(indexKey)
	// if not found then this is  _unique_.
	if kv.IsNotFound(err) {
		return nil
	}

	// no error means this is not unique
	if err == nil {
		return kv.NotUniqueError
	}

	// any other error is some sort of internal server error
	return kv.UnexpectedIndexError(err)
}

func (s *Service) uniqueLabelName(ctx context.Context, tx kv.Tx, lbl *influxdb.Label) error {
	key, err := labelIndexKey(lbl)
	if err != nil {
		return err
	}

	// labels are unique by `organization:label_name`
	err = s.unique(ctx, tx, labelIndex, key)
	if err == kv.NotUniqueError {
		return labelAlreadyExistsError(lbl)
	}
	return err
}
