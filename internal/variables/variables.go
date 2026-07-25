package variables

import (
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var varRegex = regexp.MustCompile(`\{\{([a-zA-Z0-9_.-]+)\}\}`)

type Resolver struct {
	globalVars     map[string]string
	workspaceVars  map[string]string
	collectionVars map[string]string
	requestVars    map[string]string
}

func NewResolver(global, workspace, collection, request map[string]string) *Resolver {
	if global == nil {
		global = make(map[string]string)
	}
	if workspace == nil {
		workspace = make(map[string]string)
	}
	if collection == nil {
		collection = make(map[string]string)
	}
	if request == nil {
		request = make(map[string]string)
	}
	return &Resolver{
		globalVars:     global,
		workspaceVars:  workspace,
		collectionVars: collection,
		requestVars:    request,
	}
}

func (r *Resolver) Resolve(input string) string {
	if input == "" {
		return input
	}

	return varRegex.ReplaceAllStringFunc(input, func(match string) string {
		varName := varRegex.FindStringSubmatch(match)[1]

		// Check dynamic variables first
		if dynamicVal, ok := resolveDynamicVar(varName); ok {
			return dynamicVal
		}

		// Priority order: Request -> Collection -> Workspace -> Global -> OS Env
		if val, ok := r.requestVars[varName]; ok {
			return val
		}
		if val, ok := r.collectionVars[varName]; ok {
			return val
		}
		if val, ok := r.workspaceVars[varName]; ok {
			return val
		}
		if val, ok := r.globalVars[varName]; ok {
			return val
		}
		if val := os.Getenv(varName); val != "" {
			return val
		}

		// Return original if unresolved
		return match
	})
}

func resolveDynamicVar(name string) (string, bool) {
	switch name {
	case "timestamp":
		return strconv.FormatInt(time.Now().Unix(), 10), true
	case "uuid":
		return uuid.New().String(), true
	case "randomInt":
		return strconv.Itoa(rand.Intn(10000)), true
	case "randomString":
		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 8)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		return string(b), true
	case "hostname":
		host, err := os.Hostname()
		if err != nil {
			return "localhost", true
		}
		return host, true
	default:
		return "", false
	}
}

func ReplaceInMap(r *Resolver, m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	res := make(map[string]string, len(m))
	for k, v := range m {
		res[r.Resolve(k)] = r.Resolve(v)
	}
	return res
}
