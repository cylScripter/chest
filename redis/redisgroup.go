package redisgroup

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/rpc"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

func New(arg ...*RedisNodeConfig) (g *RedisGroup, err error) {
	g = new(RedisGroup)
	if len(arg) == 0 {
		return nil, errors.New("empty redis node config")
	}
	g.mu = sync.RWMutex{}
	for _, c := range arg {

		n := newRedisNode(c.IP, c.Port, c.Password, c.Db)
		g.nodeList = append(g.nodeList, n)
		g.config = append(g.config, c)
	}
	return g, nil
}

type RedisNode struct {
	client *redis.Client
	ip     string
	port   int
	hash   uint32
	db     int
}

func hashStr(s string) uint32 {
	f := fnv.New32a()
	_, _ = f.Write([]byte(s))
	return f.Sum32()
}

type RedisNodeConfig struct {
	IP       string
	Port     int
	Password string
	Db       int
}

func newRedisNode(ip string, port int, password string, db int) *RedisNode {
	n := new(RedisNode)
	n.ip = ip
	n.port = port
	n.db = db
	n.hash = hashStr(fmt.Sprintf("%s:%d", ip, port))
	n.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", ip, port),
		Password: password,
		DB:       db,
	})
	return n
}

type RedisGroup struct {
	nodeList []*RedisNode
	config   []*RedisNodeConfig
	mu       sync.RWMutex
}

type nodeListSorter struct {
	nodeList []*RedisNode
}

func (s *nodeListSorter) Len() int {
	return len(s.nodeList)
}

func (s *nodeListSorter) Less(i, j int) bool {
	return s.nodeList[i].hash < s.nodeList[j].hash
}

func (s *nodeListSorter) Swap(i, j int) {
	tmp := s.nodeList[i]
	s.nodeList[i] = s.nodeList[j]
	s.nodeList[j] = tmp
}

func (g *RedisGroup) sortNodes() {
	i := &nodeListSorter{nodeList: g.nodeList}
	sort.Sort(i)
	g.nodeList = i.nodeList
}

func (g *RedisGroup) FindClient4Key(key string) *redis.Client {
	hash := hashStr(key)
	g.mu.RLock()
	defer g.mu.RUnlock()
	if len(g.nodeList) == 0 {
		return nil
	}
	i := 0
	for i+1 < len(g.nodeList) {
		if hash >= g.nodeList[i].hash && hash < g.nodeList[i+1].hash {
			return g.nodeList[i].client
		}
		i++
	}
	return g.nodeList[len(g.nodeList)-1].client
}

func (g *RedisGroup) GetNodes() []*RedisNode {
	return g.nodeList
}

func (g *RedisGroup) GetClient(node *RedisNode) *redis.Client {
	return node.client
}

func (g *RedisGroup) Set(key string, val []byte, exp time.Duration) error {
	log.Debugf("redis: Set: key %s exp %v", key, exp)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.Set(key, val, exp).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}

	// keyLen := len(key)
	// valueLen := len(val)
	//
	// if exp == 0 {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonNotSetExpTime,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//        ValueLen:    valueLen,
	//        Exp:         exp,
	//     })
	// }
	//
	// if keyLen > KeyLenLimit {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonKeyToLong,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//        ValueLen:    valueLen,
	//        Exp:         exp,
	//     })
	// }
	//
	// if valueLen > ValueLenLimit {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonValueToLong,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//        ValueLen:    valueLen,
	//        Exp:         exp,
	//     })
	// }

	return nil
}

func (g *RedisGroup) SetAny(key string, val interface{}) error {
	log.Debugf("redis: Set: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.Set(key, val, 0).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	return nil
}

func (g *RedisGroup) SetEx(key string, val interface{}, expire time.Duration) error {
	if expire <= 0 {
		return errors.New("expire must be greater than 0")
	}

	log.Debugf("redis: Set: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}

	buf, err := toString(val)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = node.Set(key, buf, expire).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	return nil
}

func (g *RedisGroup) SetUint64(key string, val uint64, exp time.Duration) error {
	log.Debugf("redis: Set: key %s exp %v", key, exp)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.Set(key, val, exp).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}

	// keyLen := len(key)
	//
	// if exp == 0 {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonNotSetExpTime,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//        Exp:         exp,
	//     })
	// }
	//
	// if keyLen > KeyLenLimit {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonKeyToLong,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//        Exp:         exp,
	//     })
	// }

	return nil
}

func (g *RedisGroup) SetRange(key string, offset int64, value string) error {
	log.Debugf("redis: SetRange: key %s offset %d", key, offset)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.SetRange(key, offset, value).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	return nil
}

func (g *RedisGroup) SetJson(key string, j interface{}, exp time.Duration) error {
	val, err := json.Marshal(j)
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	// 空串这里先不考虑
	if len(val) == 0 {
		return errors.New("unsupported empty value")
	}
	return g.Set(key, val, exp)
}

func (g *RedisGroup) HLen(key string) (uint32, error) {
	log.Debugf("redis: HLen: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	v := node.HLen(key)
	err := v.Err()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		log.Errorf("err:%v", err)
		return 0, err
	}
	return uint32(v.Val()), nil
}

func (g *RedisGroup) HSetJson(key, subKey string, j interface{}, exp time.Duration) error {
	log.Debugf("redis: HSetJson: key %s subKey %+v", key, subKey)
	val, err := json.Marshal(j)
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err = node.HSet(key, subKey, string(val)).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	if exp > 0 {
		err = node.Expire(key, exp).Err()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

func (g *RedisGroup) HSetJsonNX(key, subKey string, j interface{}, exp time.Duration, setSuccess *bool) error {
	log.Debugf("redis: HSetJsonNX: key %s subKey %+v exp %v", key, subKey, exp)
	val, err := json.Marshal(j)
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	res := node.HSetNX(key, subKey, string(val))
	err = res.Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	if exp > 0 {
		err = node.Expire(key, exp).Err()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	if setSuccess != nil {
		*setSuccess = res.Val()
	}
	return nil
}

func (g *RedisGroup) Get(key string) ([]byte, error) {
	log.Debugf("redis: Get: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	val, err := node.Get(key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%s", err)
		}
		return nil, err
	}
	return val, nil
}

var ErrRedisException = errors.New("redis exception")
var ErrJsonUnmarshal = rpc.CreateErrorWithMsg(6001, "json unmarshal error")

func (g *RedisGroup) GetJson(key string, j interface{}) error {
	val, err := g.Get(key)
	if err != nil {
		if err == redis.Nil {
			return redis.Nil
		}
		log.Errorf("err:%s", err)
		return ErrRedisException
	}
	err = json.Unmarshal(val, j)
	if err != nil {
		log.Errorf("err:%s", err)
		return ErrJsonUnmarshal
	}
	return nil
}

func (g *RedisGroup) GetUint64(key string) (uint64, error) {
	val, err := g.Get(key)
	if err != nil {
		if err == redis.Nil {
			return 0, redis.Nil
		}
		log.Errorf("err:%s", err)
		return 0, ErrRedisException
	}

	i, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return uint64(i), nil
}

func (g *RedisGroup) GetIntDef(key string, def int) (int, error) {
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.Get(key).Int64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%s", err)
			return def, err
		}
		return def, nil
	}

	return int(val), nil
}

func (g *RedisGroup) GetInt64(key string) (int64, error) {
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.Get(key).Int64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%s", err)
		}
		return 0, err
	}

	return val, nil
}

func (g *RedisGroup) GetInt64Def(key string, def int64) (int64, error) {
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.Get(key).Int64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%s", err)
			return def, err
		}
		return def, nil
	}

	return val, nil
}

func (g *RedisGroup) HGetAll(key string) (map[string]string, error) {
	log.Debugf("redis: HGetAll: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	return node.HGetAll(key).Result()
}

func (g *RedisGroup) HKeys(key string) ([]string, error) {
	log.Debugf("redis: HKeys: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	return node.HKeys(key).Result()
}

func (g *RedisGroup) HScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	log.Debugf("redis: HKeys: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, 0, errors.New("not found available redis node")
	}
	return node.HScan(key, cursor, match, count).Result()
}

func (g *RedisGroup) ZScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	log.Debugf("redis: ZKeys: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, 0, errors.New("not found available redis node")
	}
	return node.ZScan(key, cursor, match, count).Result()
}

func (g *RedisGroup) HMGetJson(key, subKey string, j interface{}) error {
	log.Debugf("redis: HMGetJson: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	values, err := node.HMGet(key, subKey).Result()
	if err != nil {
		log.Errorf("redis HMGet err:%v", err)
		return ErrRedisException
	}
	if len(values) == 1 {
		v := values[0]
		if v != nil {
			var buf []byte
			if p, ok := v.(string); ok {
				buf = []byte(p)
			} else if p, ok := v.([]byte); ok {
				buf = p
			}
			if buf != nil {
				if len(buf) > 0 {
					err = json.Unmarshal(buf, j)
					if err != nil {
						log.Errorf("err:%s", err)
						return ErrJsonUnmarshal
					}
				}
				return nil
			}
		}
	}
	return redis.Nil
}

// errorKeyList  返回有序列化问题的key
func (g *RedisGroup) HMBatchGetJson(key string, m map[string]interface{}) ([]string, error) {
	log.Debugf("redis: HMBatchGetJson: key %s", key)
	if len(m) == 0 {
		return nil, nil
	}
	var subKeys []string
	for k := range m {
		if k != "" {
			subKeys = append(subKeys, k)
		}
	}
	if len(subKeys) == 0 {
		return nil, nil
	}
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	values, err := node.HMGet(key, subKeys...).Result()
	if err != nil {
		log.Errorf("redis HMGet err:%v", err)
		return nil, ErrRedisException
	}
	var errUnmarshalKeyList []string
	for i, v := range values {
		if v != nil {
			var buf []byte
			if p, ok := v.(string); ok {
				buf = []byte(p)
			} else if p, ok := v.([]byte); ok {
				buf = p
			}
			if buf != nil {
				if len(buf) > 0 && i < len(subKeys) {
					subKey := subKeys[i]
					j := m[subKey]
					if j != nil {
						err = json.Unmarshal(buf, j)
						if err != nil {
							log.Errorf("err:%s", err)
							errUnmarshalKeyList = append(errUnmarshalKeyList, subKey)
						}
					}
				}
			}
		}
	}
	if len(errUnmarshalKeyList) > 0 {
		return errUnmarshalKeyList, ErrJsonUnmarshal
	}
	return nil, nil
}

func (g *RedisGroup) HMBatchGet(key string, subKeys []string) ([]interface{}, error) {
	log.Debugf("redis: HMBatchGet: key %s", key)
	if len(subKeys) == 0 {
		return nil, nil
	}
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	values, err := node.HMGet(key, subKeys...).Result()
	if err != nil {
		log.Errorf("redis HMBatchGet err:%v", err)
		return nil, err
	}
	return values, nil
}

func (g *RedisGroup) HDel(key string, subKey ...string) (int64, error) {
	log.Debugf("redis: HDel: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	delNum, err := node.HDel(key, subKey...).Result()
	if err != nil {
		log.Errorf("err:%s", err)
		return 0, err
	}
	return delNum, nil
}

func (g *RedisGroup) Del(key string) error {
	log.Debugf("redis: Del: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.Del(key).Err()
	if err != nil {
		log.Errorf("err:%s", err)
		return err
	}
	return nil
}

func (g *RedisGroup) ZAdd(key string, values ...redis.Z) error {
	log.Debugf("redis: ZAdd: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.ZAdd(key, values...).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (g *RedisGroup) ZCount(key, min, max string) (int64, error) {
	log.Debugf("redis: ZCount: key %s min %s - max %s", key, min, max)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	v := node.ZCount(key, min, max)
	err := v.Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return v.Val(), nil
}

func (g *RedisGroup) ZRangeByScore(key string, opt redis.ZRangeBy) ([]string, error) {
	log.Debugf("redis: ZRangeByScore: key %s opt %v", key, opt)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	members, err := node.ZRangeByScore(key, opt).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return members, nil
}

func (g *RedisGroup) ZRangeByScoreWithScores(key string, opt redis.ZRangeBy) ([]redis.Z, error) {
	log.Debugf("redis: ZRangeByScoreWithScores: key %s opt %v", key, opt)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	resultList, err := node.ZRangeByScoreWithScores(key, opt).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return resultList, nil
}

func (g *RedisGroup) ZIncrBy(key string, increment float64, member string) error {
	log.Debugf("redis: ZIncrBy: key %s increment %v member %s", key, increment, member)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	_, err := node.ZIncrBy(key, increment, member).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (g *RedisGroup) ZRange(key string, start, stop int64) ([]string, error) {
	log.Debugf("redis: ZRange: key %s start %v stop %v", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	members, err := node.ZRange(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return members, nil
}

func (g *RedisGroup) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	log.Debugf("redis: ZRangeWithScores: key %s start %v stop %v", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	resultList, err := node.ZRangeWithScores(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return resultList, nil
}

func (g *RedisGroup) ZRevRange(key string, start, stop int64) ([]string, error) {
	log.Debugf("redis: ZRevRange: key %s start %v stop %v", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	members, err := node.ZRevRange(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return members, nil
}

func (g *RedisGroup) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	log.Debugf("redis: ZRevRangeWithScores: key %s start %v stop %v", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	resultList, err := node.ZRevRangeWithScores(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return resultList, nil
}

func (g *RedisGroup) ZRem(key string, members ...interface{}) (int64, error) {
	log.Debugf("redis: ZRem: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	delNum, err := node.ZRem(key, members...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return delNum, nil
}

func (g *RedisGroup) ZRemRangeByScore(key string, min, max string) (int64, error) {
	log.Debugf("redis: ZRemRangeByScore: key %s, min: %s, max: %s", key, min, max)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	delNum, err := node.ZRemRangeByScore(key, min, max).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return delNum, nil
}

func (g *RedisGroup) ZCard(key string) (int64, error) {
	log.Debugf("redis: ZCard: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	num, err := node.ZCard(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return num, nil
}

func (g *RedisGroup) ZRank(key string, member string) (int64, error) {
	log.Debugf("redis: ZRank: key %s member %v", key, member)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	idx, err := node.ZRank(key, member).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return idx, nil
}

func (g *RedisGroup) SAdd(key string, values ...interface{}) error {
	log.Debugf("redis: SAdd: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.SAdd(key, values...).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (g *RedisGroup) SRem(key string, members ...interface{}) (int64, error) {
	log.Debugf("redis: SRem: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	delNum, err := node.SRem(key, members...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return delNum, nil
}

func (g *RedisGroup) SCard(key string) (int64, error) {
	log.Debugf("redis: SCard: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	num, err := node.SCard(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return num, nil
}

func (g *RedisGroup) SIsMember(key string, members interface{}) (bool, error) {
	log.Debugf("redis: SIsMember: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return false, errors.New("not found available redis node")
	}
	ok, err := node.SIsMember(key, members).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (g *RedisGroup) SMembers(key string) ([]string, error) {
	log.Debugf("redis: SMembers: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	members, err := node.SMembers(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return members, nil
}

func (g *RedisGroup) HIncrBy(key, field string, incr int64) (int64, error) {
	log.Debugf("redis: HIncrBy: key %s field %s incr %d", key, field, incr)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	n, err := node.HIncrBy(key, field, incr).Result()
	if err != nil {
		log.Errorf("err:%s", err)
		return 0, err
	}
	return n, nil
}

func (g *RedisGroup) HIncr(key, field string) (int64, error) {
	log.Debugf("redis: HIncrBy: key %s field %s incr %d", key, field, 1)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	n, err := node.HIncrBy(key, field, 1).Result()
	if err != nil {
		log.Errorf("err:%s", err)
		return 0, err
	}
	return n, nil
}

func (g *RedisGroup) IncrBy(key string, incr int64) (int64, error) {
	log.Debugf("redis: IncrBy: key %s incr %d", key, incr)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	n, err := node.IncrBy(key, incr).Result()
	if err != nil {
		log.Errorf("err:%s", err)
		return 0, err
	}
	return n, nil
}

func (g *RedisGroup) DecrBy(key string, decr int64) (int64, error) {
	log.Debugf("redis: DecrBy: key %s decr %d", key, decr)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	n, err := node.DecrBy(key, decr).Result()
	if err != nil {
		log.Errorf("err:%s", err)
		return 0, err
	}
	return n, nil
}

func (g *RedisGroup) HSet(key, field string, val interface{}) error {
	log.Debugf("redis: HSet: key %s field %s", key, field)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.HSet(key, field, val).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (g *RedisGroup) HMSet(key string, fields map[string]interface{}) error {
	log.Debugf("redis: HMSet: key %s fields %v", key, fields)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	err := node.HMSet(key, fields).Err()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (g *RedisGroup) HGet(key, subKey string) (string, error) {
	log.Debugf("redis: HGet: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	val, err := node.HGet(key, subKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%v", err)
		}
		return "", err
	}
	return val, nil
}

func (g *RedisGroup) HGetUint64(key, subKey string) (uint64, error) {
	log.Debugf("redis: HGet: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.HGet(key, subKey).Uint64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%v", err)
		}
		return 0, err
	}
	return val, nil
}

func (g *RedisGroup) HGetIntDef(key, subKey string, def int) (int, error) {
	log.Debugf("redis: HGet: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.HGet(key, subKey).Int64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%v", err)
			return def, err
		}
		return def, nil
	}
	return int(val), nil
}

func (g *RedisGroup) HGetInt(key, subKey string) (int, error) {
	log.Debugf("redis: HGet: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.HGet(key, subKey).Int64()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%v", err)
		}
		return 0, err
	}
	return int(val), nil
}

func (g *RedisGroup) HGetJson(key, subKey string, j interface{}) error {
	log.Debugf("redis: HGetJson: key %s subKey %+v", key, subKey)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	val, err := node.HGet(key, subKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Errorf("err:%v", err)
		}
		return err
	}
	err = json.Unmarshal([]byte(val), j)
	if err != nil {
		log.Errorf("err:%s", err)
		return ErrJsonUnmarshal
	}
	return nil
}

func (g *RedisGroup) Expire(key string, expiration time.Duration) error {
	log.Debugf("redis: Expire: key %s exp %+v", key, expiration)
	node := g.FindClient4Key(key)
	if node == nil {
		return errors.New("not found available redis node")
	}
	_, err := node.Expire(key, expiration).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (g *RedisGroup) Exists(key string) (bool, error) {
	log.Debugf("redis: Exists: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return false, errors.New("not found available redis node")
	}
	val, err := node.Exists(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	if val == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (g *RedisGroup) HExists(key, field string) (bool, error) {
	log.Debugf("redis: HExists: key %s field %s", key, field)
	node := g.FindClient4Key(key)
	if node == nil {
		return false, errors.New("not found available redis node")
	}
	exists, err := node.HExists(key, field).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return exists, nil
}

func (g *RedisGroup) ScriptRun(lua string, keys []string, args ...interface{}) (interface{}, error) {
	node := g.FindClient4Key(keys[0])
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	script := redis.NewScript(lua)
	result, err := script.Run(node, keys, args...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return result, nil
}

func (g *RedisGroup) EvalSha(luaSha1 string, keys []string, args ...interface{}) (interface{}, error) {
	node := g.FindClient4Key(keys[0])
	if node == nil {
		return nil, errors.New("not found available redis node")
	}
	result, err := node.EvalSha(luaSha1, keys, args...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return result, nil
}

func (g *RedisGroup) ScriptLoad(luaScript string) (string, error) {
	node := g.FindClient4Key("")
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	luaSha1, err := node.ScriptLoad(luaScript).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return luaSha1, nil
}

func (g *RedisGroup) Incr(key string) (int64, error) {
	log.Debugf("redis: Incr: key %s", key)
	node := g.FindClient4Key("")
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.Incr(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (g *RedisGroup) Decr(key string) (int64, error) {
	log.Debugf("redis: Decr: key %s", key)
	node := g.FindClient4Key("")
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	val, err := node.Decr(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return val, nil
}

func (g *RedisGroup) ExpireAt(key string, expiredAt time.Time) (bool, error) {
	log.Debugf("redis: ExpireAt: key %s exp %v", key, expiredAt)
	node := g.FindClient4Key(key)
	if node == nil {
		return false, errors.New("not found available redis node")
	}
	ok, err := node.ExpireAt(key, expiredAt).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return ok, nil
}

func (g *RedisGroup) LPop(key string) (string, error) {
	log.Debugf("redis: LPop: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	val, err := node.LPop(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (g *RedisGroup) RPop(key string) (string, error) {
	log.Debugf("redis: RPop: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	val, err := node.RPop(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (g *RedisGroup) LPush(key string, values ...interface{}) (int64, error) {
	log.Debugf("redis: LPush: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	count, err := node.LPush(key, values...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return count, nil
}

func (g *RedisGroup) RPush(key string, values ...interface{}) (int64, error) {
	log.Debugf("redis: RPush: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	count, err := node.RPush(key, values...).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return count, nil
}

func (g *RedisGroup) LRange(key string, start, stop int64) ([]string, error) {
	log.Debugf("redis: LRange: key %s start %d stop %d", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return []string{}, errors.New("not found available redis node")
	}
	result, err := node.LRange(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return []string{}, err
	}
	return result, nil
}

func (g *RedisGroup) LTrim(key string, start, stop int64) (string, error) {
	log.Debugf("redis: LTrim: key %s start %d stop %d", key, start, stop)
	node := g.FindClient4Key(key)
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	result, err := node.LTrim(key, start, stop).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return result, nil
}

func (g *RedisGroup) LLen(key string) (int64, error) {
	log.Debugf("redis: LLen: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	count, err := node.LLen(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return count, nil
}

func (g *RedisGroup) LIndex(key string, index int64) (string, error) {
	log.Debugf("redis: LIndex: key %s index %d", key, index)
	node := g.FindClient4Key(key)
	if node == nil {
		return "", errors.New("not found available redis node")
	}
	val, err := node.LIndex(key, index).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}
	return val, nil
}

func (g *RedisGroup) SetNX(key string, val []byte, exp time.Duration) (bool, error) {
	log.Debugf("redis: SetNX: key %s exp %v", key, exp)
	node := g.FindClient4Key(key)
	if node == nil {
		return false, errors.New("not found available redis node")
	}
	b, err := node.SetNX(key, val, exp).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}
	return b, nil
}

func (g *RedisGroup) ZScore(key string, member string) (float64, error) {
	log.Debugf("redis: ZScore: key %s member %s", key, member)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	score, err := node.ZScore(key, member).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return score, nil
}

func (g *RedisGroup) Ttl(key string) (time.Duration, error) {
	log.Debugf("redis: TTL: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	ttl, err := node.TTL(key).Result()
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}
	return ttl, nil
}

func (g *RedisGroup) SetBit(key string, offset int64, val int) (int64, error) {
	log.Debugf("redis: Set Bit: key %s offset %d val %d", key, offset, val)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	intCmd := node.SetBit(key, offset, val)
	if intCmd.Err() != nil {
		log.Errorf("err:%s", intCmd.Err())
		return 0, intCmd.Err()
	}

	// keyLen := len(key)
	// if keyLen > KeyLenLimit {
	//     stream.LogStreamSubType(nil, subType, &RedisExceptionMsg{
	//        ServiceName: rpc.ServiceName,
	//        Reason:      ReasonKeyToLong,
	//        Key:         key,
	//        KeyLen:      keyLen,
	//     })
	// }
	return intCmd.Result()
}

func (g *RedisGroup) GetBit(key string, offset int64) (int64, error) {
	log.Debugf("redis: Set Bit: key %s offset %d", key, offset)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	intCmd := node.GetBit(key, offset)
	if intCmd.Err() != nil {
		log.Errorf("err:%s", intCmd.Err())
		return 0, intCmd.Err()
	}
	return intCmd.Result()
}

func (g *RedisGroup) BitCount(key string, bitCount *redis.BitCount) (int64, error) {
	log.Debugf("redis: BitCount: key %s", key)
	node := g.FindClient4Key(key)
	if node == nil {
		return 0, errors.New("not found available redis node")
	}
	intCmd := node.BitCount(key, bitCount)
	if intCmd.Err() != nil {
		log.Errorf("err:%s", intCmd.Err())
		return 0, intCmd.Err()
	}
	return intCmd.Result()
}

func (g *RedisGroup) FlushAll() (string, error) {
	node := g.FindClient4Key("")
	if node == nil {
		return "", errors.New("not found available redis node")
	}

	intCmd := node.FlushAll()
	if intCmd.Err() != nil {
		log.Errorf("err:%s", intCmd.Err())
		return "", intCmd.Err()
	}
	return intCmd.Result()
}

func (g *RedisGroup) IsNotFound(err error) bool {
	return errors.Is(err, redis.Nil)
}
