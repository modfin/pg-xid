package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/modfin/henry/slicez"
	"github.com/rs/xid"
	"io/ioutil"
	"testing"
	"time"
)

func startpg() types.Container {
	fmt.Println("Starting postgres")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	fmt.Println("Pulling image")
	_, err = cli.ImagePull(ctx, "postgres", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	//defer reader.Close()
	fmt.Println("Creating container")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "postgres",
		ExposedPorts: map[nat.Port]struct{}{"5432/tcp": {}},
		Env:          []string{"POSTGRES_PASSWORD=qwerty"},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "5432",
				},
			},
		},
	}, nil, nil, "xid_test")
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	for _, c := range containers {
		if c.ID == resp.ID {
			return c
		}
	}
	panic("could not find container")

}

func stoppg(c types.Container) {
	fmt.Println("Stopping postgres")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	timeout := 5 * time.Second
	err = cli.ContainerStop(ctx, c.ID, &timeout)
	if err != nil {
		panic(err)
	}
	fmt.Println("Removing postgres")
	err = cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}

func waitForDb() (*sqlx.DB, error) {
	cuttoff := time.Now().Add(10 * time.Second)
	for time.Now().Before(cuttoff) {
		time.Sleep(time.Second / 2)
		db, err := sqlx.Open("postgres", "postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable")
		if err != nil {
			//fmt.Println(1, err)
			continue
		}
		_, err = db.Exec("SELECT 1")
		if err != nil {
			//fmt.Println(2, err)
			continue
		}
		return db, nil
	}
	return nil, errors.New("could not connect to database withing 10 sec")

}

func installXid(db *sqlx.DB) {
	fmt.Println("Install ./xid.sql")
	b, err := ioutil.ReadFile("./xid.sql")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(string(b))
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("SELECT setval('xid_serial', 16777215/2);")
	if err != nil {
		panic(err)
	}
}

func TestXid(t *testing.T) {
	c := startpg()
	defer stoppg(c)
	db, err := waitForDb()
	if err != nil {
		panic(err)
	}
	installXid(db)

	t.Run("Generate", generate(db))
	t.Run("GenerateAt", generateAt(db))
	t.Run("Encoding", encoding(db))
	t.Run("Decoding", decoding(db))
	t.Run("Inspection", inspection(db))
	t.Run("CyclingCounter", testCycle(db))
}

func testCycle(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		off := 100
		max := 16777215 + 1
		start := max - off

		_, err := db.Exec("SELECT setval('xid_serial', $1);", start-1)
		if err != nil {
			t.Log("unexpected err, got", err)
			t.Fail()
		}

		var xs []string
		err = db.Select(&xs, "SELECT xid() FROM generate_series(1,$1)", off*2)
		if err != nil {
			t.Log("unexpected err, got", err)
			t.Fail()
		}

		if len(xs) != off*2 {
			t.Log("expected len ==", off*2, "got", len(xs))
			t.Fail()
		}

		for i, str := range xs {
			x, err := xid.FromString(str)
			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}
			counter := x.Counter()

			if counter != int32((start+i)%max) {
				t.Logf("expected counter %v, got %v", int32((start+i)%max), counter)
				t.Fail()
			}

		}

	}
}

func inspection(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("Time", inspectTime(db))
		t.Run("Counter", inspectCounter(db))
		t.Run("Pid", inspectPid(db))
		t.Run("Machine", inspectMachine(db))

	}
}

func inspectMachine(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {
		for i := 0; i < 10; i++ {
			x := xid.New()
			var got pq.Int32Array
			err := db.Get(&got, "SELECT xid_machine($1)", x.String())

			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			exp := slicez.Map(x.Machine(), func(a byte) int32 {
				return int32(a)
			})

			if !slicez.Equal(exp, got) {
				t.Logf("expected %v, got %v", exp, got)
				t.Fail()
			}
		}
	}
}

func inspectPid(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {
		for i := 0; i < 10; i++ {
			x := xid.New()
			var got uint16
			err := db.Get(&got, "SELECT xid_pid($1)", x.String())

			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			exp := x.Pid()
			if exp != got {
				t.Logf("expected %v, got %v", exp, got)
				t.Fail()
			}
		}
	}
}

func inspectCounter(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		for i := 0; i < 100; i++ {
			x := xid.New()
			var got int32
			err := db.Get(&got, "SELECT xid_counter($1)", x.String())

			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			exp := x.Counter()
			if exp != got {
				t.Logf("expected %v, got %v", exp, got)
				t.Fail()
			}
		}
	}
}

func inspectTime(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		for i := 0; i < 100; i++ {
			refTime := time.Now().
				Add(time.Duration(-i) * time.Millisecond).
				Add(time.Duration(-i) * time.Hour).
				Add(time.Duration(-i) * time.Minute).
				Add(time.Duration(-i) * time.Second)

			x := xid.NewWithTime(refTime)
			var got time.Time
			err := db.Get(&got, "SELECT xid_time($1)", x.String())

			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			exp := x.Time()
			if !exp.Equal(got) {
				t.Logf("expected %s, got %s", exp, got)
				t.Fail()
			}
		}
	}
}

func encoding(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		for i := 0; i < 100; i++ {
			x := xid.New()
			b := slicez.Map(x.Bytes(), func(b byte) int32 { return int32(b) })

			var str string
			err := db.Get(&str, "SELECT xid_encode($1)", pq.Int32Array(b))
			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			if str != x.String() {
				t.Logf("expected %s, got %s", x.String(), str)
				t.Fail()
			}

		}

	}
}

func decoding(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		for i := 0; i < 100; i++ {
			x := xid.New()

			str := x.String()

			var res pq.Int32Array
			err := db.Get(&res, "SELECT xid_decode($1)", str)
			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			b := slicez.Map(x.Bytes(), func(b byte) int32 { return int32(b) })

			if !slicez.Equal(b, res) {
				t.Logf("expected %v, got %v", b, str)
				t.Fail()
			}

		}

	}
}

func generate(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		var ids []string

		start := time.Now().Add(-time.Second)
		err := db.Select(&ids, "SELECT xid() FROM generate_series(1, 100)")
		if err != nil {
			t.Log("unexpected err, got", err)
			t.Fail()
		}
		done := time.Now()

		var last xid.ID
		for i, id := range ids {
			x, err := xid.FromString(id)
			if err != nil {
				t.Log("expected a valid xid, got", err)
				t.Fail()
			}
			if !start.Before(x.Time()) {
				t.Logf("expected xid to contain timestamp after %v, but it held %v", start, x.Time())
				t.Fail()
			}
			if !x.Time().Before(done) {
				t.Logf("expected xid to contain timestamp before %v, but it held %v", done, x.Time())
				t.Fail()
			}
			if i > 0 {
				if last.Counter()+1 != x.Counter() {
					t.Logf("expected xid counter to be incremented by 1, last %v, but it held %v", last.Counter(), x.Counter())
					t.Fail()
				}
			}
			last = x
		}

	}
}

func generateAt(db *sqlx.DB) func(t *testing.T) {
	return func(t *testing.T) {

		var last xid.ID
		for i := 0; i < 1000; i++ {
			refTime := time.Now().
				Add(time.Duration(-i*3) * time.Millisecond).
				Add(time.Duration(-i) * time.Hour).
				Add(time.Duration(-i) * time.Minute).
				Add(time.Duration(-i) * time.Second)

			var str string
			err := db.Get(&str, "SELECT xid(_at => $1)", refTime)

			if err != nil {
				t.Log("unexpected err, got", err)
				t.Fail()
			}

			expTime := xid.NewWithTime(refTime).Time()
			got, err := xid.FromString(str)
			if err != nil {
				t.Log("expected a valid xid, got", err, str)
				t.Fail()
				return
			}

			if !expTime.Equal(got.Time()) {
				t.Logf("expected %s, got %s", expTime, got.Time())
				t.Fail()
			}

			if i > 0 {
				if last.Counter()+1 != got.Counter() {
					t.Logf("expected xid counter to be incremented by 1, last %v, but it held %v", last.Counter(), got.Counter())
					t.Fail()
				}
			}
			last = got
		}

	}
}
