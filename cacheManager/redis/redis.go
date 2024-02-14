package redis

/*
	ref - https://redis.io/docs/connect/clients/go/
	Author - Akshay.Verma
	Github - github.com/Akshay-Verma-CS
*/

import (
	"auth-go/Constants"
	"auth-go/configuration"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client        *redis.Client
	clusterClient *redis.ClusterClient
	once          sync.Once
	mutex         sync.Mutex
)

var opt *redis.Options

var redisConfig = configuration.GetConfig().Cache.Redis

func Connect() {
	if redisConfig.Url != "" {
		var err error
		opt, err = redis.ParseURL(redisConfig.Url)

		if err != nil {
			fmt.Println("Error occurred while parsing redis url")
		}
	} else {
		opt = &redis.Options{
			Addr:     redisConfig.Address,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		}
	}

	client = redis.NewClient(opt)
}

func GetClient() *redis.Client {
	mutex.Lock()
	defer mutex.Unlock()

	if client == nil {
		once.Do(Connect)
	} else {
		if err := client.Ping(context.Background()).Err(); err != nil {
			fmt.Println("Redis connection lost. Attempting to reconnect...")
			once = sync.Once{} // Reset once to allow re-initialization
			once.Do(Connect)
		}
	}
	return client
}

func ConnectToCluster() {
	clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: redisConfig.ClusterAddress,
	})
}

func GetClusterClient() *redis.ClusterClient {
	mutex.Lock()
	defer mutex.Unlock()

	if clusterClient == nil {
		once.Do(ConnectToCluster)
	} else {
		if err := clusterClient.Ping(context.Background()).Err(); err != nil {
			fmt.Println("Redis connection lost. Attempting to reconnect...")
			once = sync.Once{}
			once.Do(ConnectToCluster)
		}
	}
	return clusterClient
}

func ConnectWithTLS() {
	cert, err := tls.LoadX509KeyPair(Constants.REDIS_USER_CRT_PATH, Constants.REDIS_USER_PRIVATE_KEY_PATH)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := os.ReadFile(Constants.REDIS_USER_PEM_FILE_PATH)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address,
		Username: redisConfig.Username, //More info https://redis.io/docs/management/security/acl/
		Password: redisConfig.Password,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	})
}

func Store(ctx context.Context, key string, value interface{}) {
	err := client.Set(ctx, key, value, time.Duration(redisConfig.Expiration))
	if err != nil {
		panic(err)
	}
}

func Retrieve(ctx context.Context, key string) interface{} {
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	return val
}

func StoreMap(ctx context.Context, mapName string, Map map[string]string) {
	for k, v := range Map {
		err := client.HSet(ctx, mapName, k, v).Err()
		if err != nil {
			panic(err)
		}
	}
}

func RetrieveMap(ctx context.Context, mapName string) map[string]string {
	return client.HGetAll(ctx, mapName).Val()
}
