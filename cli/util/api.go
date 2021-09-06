package cliutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/node/config"
	"github.com/filecoin-project/lotus/node/repo"
)

const (
	metadataTraceContext = "traceContext"
)

// flagsForAPI returns flags passed on the command line with the listen address
// of the API server (only used by the tests), in the order of precedence they
// should be applied for the requested kind of node.
func flagsForAPI(t repo.RepoType) []string {
	switch t {
	case repo.FullNode:
		return []string{"api-url"}
	case repo.StorageMiner:
		return []string{"miner-api-url"}
	case repo.Worker:
		return []string{"worker-api-url"}
	case repo.Markets:
		// support split markets-miner and monolith deployments.
		return []string{"markets-api-url", "miner-api-url"}
	default:
		panic(fmt.Sprintf("Unknown repo type: %v", t))
	}
}

func flagsForRepo(t repo.RepoType) []string {
	switch t {
	case repo.FullNode:
		return []string{"repo"}
	case repo.StorageMiner:
		return []string{"miner-repo"}
	case repo.Worker:
		return []string{"worker-repo"}
	case repo.Markets:
		// support split markets-miner and monolith deployments.
		return []string{"markets-repo", "miner-repo"}
	default:
		panic(fmt.Sprintf("Unknown repo type: %v", t))
	}
}

// EnvsForAPIInfos returns the environment variables to use in order of precedence
// to determine the API endpoint of the specified node type.
//
// It returns the current variables and deprecated ones separately, so that
// the user can log a warning when deprecated ones are found to be in use.
func EnvsForAPIInfos(t repo.RepoType) (primary string, fallbacks []string, deprecated []string) {
	switch t {
	case repo.FullNode:
		return "FULLNODE_API_INFO", nil, nil
	case repo.StorageMiner:
		// TODO remove deprecated deprecation period
		return "MINER_API_INFO", nil, []string{"STORAGE_API_INFO"}
	case repo.Worker:
		return "WORKER_API_INFO", nil, nil
	case repo.Markets:
		// support split markets-miner and monolith deployments.
		return "MARKETS_API_INFO", []string{"MINER_API_INFO"}, nil
	default:
		panic(fmt.Sprintf("Unknown repo type: %v", t))
	}
}

func getAPIInfoFromService(path string, rcfg config.Service, t repo.RepoType, checkRepoDir bool) (APIInfo, error) {
	if checkRepoDir {
		switch t {
		case repo.FullNode:
			if rcfg.Endpoints.FullNode != "" {
				return APIInfo{Addr: rcfg.Endpoints.FullNode}, nil
			}
			if p := rcfg.Endpoints.FullNodeDir; p != "" {
				path = p
			}
			return apiInfoFromRepoDir(path)
		case repo.StorageMiner:
			if rcfg.Endpoints.StorageMiner != "" {
				return APIInfo{Addr: rcfg.Endpoints.StorageMiner}, nil
			}
			if p := rcfg.Endpoints.StorageMinerDir; p != "" {
				path = p
			}
			return apiInfoFromRepoDir(path)
		case repo.Worker:
			if rcfg.Endpoints.Worker != "" {
				return APIInfo{Addr: rcfg.Endpoints.Worker}, nil
			}
			if p := rcfg.Endpoints.WorkerDir; p != "" {
				path = p
			}
			return apiInfoFromRepoDir(path)
		case repo.Markets:
			if rcfg.Endpoints.Markets != "" {
				return APIInfo{Addr: rcfg.Endpoints.Markets}, nil
			}
			if p := rcfg.Endpoints.MarketsDir; p != "" {
				path = p
			}
			return apiInfoFromRepoDir(path)
		default:
			panic(fmt.Sprintf("Unknown repo type: %v", t))
		}
	} else {
		switch t {
		case repo.FullNode:
			if rcfg.Endpoints.FullNode != "" {
				return APIInfo{Addr: rcfg.Endpoints.FullNode}, nil
			}
			if p := rcfg.Endpoints.FullNodeDir; p != "" {
				return apiInfoFromRepoDir(p)
			}
		case repo.StorageMiner:
			if rcfg.Endpoints.StorageMiner != "" {
				return APIInfo{Addr: rcfg.Endpoints.StorageMiner}, nil
			}
			if p := rcfg.Endpoints.StorageMinerDir; p != "" {
				return apiInfoFromRepoDir(p)
			}
		case repo.Worker:
			if rcfg.Endpoints.Worker != "" {
				return APIInfo{Addr: rcfg.Endpoints.Worker}, nil
			}
			if p := rcfg.Endpoints.WorkerDir; p != "" {
				return apiInfoFromRepoDir(p)
			}
		case repo.Markets:
			if rcfg.Endpoints.Markets != "" {
				return APIInfo{Addr: rcfg.Endpoints.Markets}, nil
			}
			if p := rcfg.Endpoints.MarketsDir; p != "" {
				return apiInfoFromRepoDir(p)
			}
		default:
			panic(fmt.Sprintf("Unknown repo type: %v", t))
		}

		return APIInfo{}, errors.New("missing config in rcfg")
	}
}

func tryLotusctlTOML(ctx *cli.Context, t repo.RepoType) (APIInfo, error, bool) {
	path := "~/lotusctl.toml"

	p, err := homedir.Expand(path)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("could not expand home dir (%s): %w", path, err), false
	}

	cfg := &config.Service{}
	ff, err := config.FromFile(p, cfg)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("couldn't load service config: %w", err), false
	}

	rcfg := *ff.(*config.Service)

	apiinfo, err := getAPIInfoFromService(path, rcfg, t, false)
	return apiinfo, err, true
}

func tryServiceRepoFlag(ctx *cli.Context, t repo.RepoType) (APIInfo, error) {
	flag := "service-repo"
	if !ctx.IsSet(flag) {
		return APIInfo{}, errors.New("--service-repo must be specified")
	}

	path := ctx.String(flag)

	p, err := homedir.Expand(path)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("could not expand home dir (%s): %w", path, err)
	}

	cfg := &config.Service{}
	ff, err := config.FromFile(p+"/config.toml", cfg)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("couldn't load service config: %w", err)
	}

	rcfg := *ff.(*config.Service)

	return getAPIInfoFromService(path, rcfg, t, true)
}

// GetAPIInfo returns the API endpoint to use for the specified kind of repo.
//
// The order of precedence is as follows:
//
//  1. *-api-url command line flags.
//  2. *_API_INFO environment variables
//  3. deprecated *_API_INFO environment variables
//  4. *-repo command line flags.
func GetAPIInfo(ctx *cli.Context, t repo.RepoType) (APIInfo, error) {
	// Check ~/lotusctl.toml in the HOME directory. If this file exists, it has precedence over every other deprecated flag,
	// such as --repo or --miner-repo, as well as environment variable, such as LOTUS_MINER_PATH or LOTUS_MARKETS_PATH.
	if apiinfo, err, ok := tryLotusctlTOML(ctx, t); ok {
		return apiinfo, err
	}

	// Check the config.toml within the `service-repo` directory. If this file exists, it has precedence over every other deprecated flag,
	// such as --repo or --miner-repo, as well as environment variable, such as LOTUS_MINER_PATH or LOTUS_MARKETS_PATH.
	return tryServiceRepoFlag(ctx, t)
}

func apiInfoFromRepoDir(path string) (APIInfo, error) {
	p, err := homedir.Expand(path)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("could not expand home dir (%s): %w", path, err)
	}

	r, err := repo.NewFS(p)
	if err != nil {
		return APIInfo{}, xerrors.Errorf("could not open repo at path: %s; %w", p, err)
	}

	exists, err := r.Exists()
	if err != nil {
		return APIInfo{}, xerrors.Errorf("repo.Exists returned an error: %w", err)
	}

	if !exists {
		return APIInfo{}, xerrors.Errorf("repo directory does not exist at %s", path)
	}

	ma, err := r.APIEndpoint()
	if err != nil {
		return APIInfo{}, xerrors.Errorf("could not get api endpoint: %w", err)
	}

	token, err := r.APIToken()
	if err != nil {
		log.Warnf("Couldn't load CLI token, capabilities may be limited: %v", err)
	}

	return APIInfo{Addr: ma.String(), Token: token}, nil
}

func GetRawAPI(ctx *cli.Context, t repo.RepoType, version string) (string, http.Header, error) {
	ainfo, err := GetAPIInfo(ctx, t)
	if err != nil {
		return "", nil, xerrors.Errorf("could not get API info for %s: %w", t, err)
	}

	addr, err := ainfo.DialArgs(version)
	if err != nil {
		return "", nil, xerrors.Errorf("could not get DialArgs: %w", err)
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintf(ctx.App.Writer, "using raw API %s endpoint: %s\n", version, addr)
	}

	return addr, ainfo.AuthHeader(), nil
}

func GetCommonAPI(ctx *cli.Context) (api.CommonNet, jsonrpc.ClientCloser, error) {
	ti, ok := ctx.App.Metadata["repoType"]
	if !ok {
		log.Errorf("unknown repo type, are you sure you want to use GetCommonAPI?")
		ti = repo.FullNode
	}
	t, ok := ti.(repo.RepoType)
	if !ok {
		log.Errorf("repoType type does not match the type of repo.RepoType")
	}

	if tn, ok := ctx.App.Metadata["testnode-storage"]; ok {
		return tn.(api.StorageMiner), func() {}, nil
	}
	if tn, ok := ctx.App.Metadata["testnode-full"]; ok {
		return tn.(api.FullNode), func() {}, nil
	}

	addr, headers, err := GetRawAPI(ctx, t, "v0")
	if err != nil {
		return nil, nil, err
	}

	return client.NewCommonRPCV0(ctx.Context, addr, headers)
}

func GetFullNodeAPI(ctx *cli.Context) (v0api.FullNode, jsonrpc.ClientCloser, error) {
	if tn, ok := ctx.App.Metadata["testnode-full"]; ok {
		return &v0api.WrapperV1Full{FullNode: tn.(v1api.FullNode)}, func() {}, nil
	}

	addr, headers, err := GetRawAPI(ctx, repo.FullNode, "v0")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using full node API v0 endpoint:", addr)
	}

	return client.NewFullNodeRPCV0(ctx.Context, addr, headers)
}

func GetFullNodeAPIV1(ctx *cli.Context) (v1api.FullNode, jsonrpc.ClientCloser, error) {
	if tn, ok := ctx.App.Metadata["testnode-full"]; ok {
		return tn.(v1api.FullNode), func() {}, nil
	}

	addr, headers, err := GetRawAPI(ctx, repo.FullNode, "v1")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using full node API v1 endpoint:", addr)
	}

	v1API, closer, err := client.NewFullNodeRPCV1(ctx.Context, addr, headers)
	if err != nil {
		return nil, nil, err
	}

	v, err := v1API.Version(ctx.Context)
	if err != nil {
		return nil, nil, err
	}
	if !v.APIVersion.EqMajorMinor(api.FullAPIVersion1) {
		return nil, nil, xerrors.Errorf("Remote API version didn't match (expected %s, remote %s)", api.FullAPIVersion1, v.APIVersion)
	}
	return v1API, closer, nil
}

type GetStorageMinerOptions struct {
	PreferHttp bool
}

type GetStorageMinerOption func(*GetStorageMinerOptions)

func StorageMinerUseHttp(opts *GetStorageMinerOptions) {
	opts.PreferHttp = true
}

func GetStorageMinerAPI(ctx *cli.Context, opts ...GetStorageMinerOption) (api.StorageMiner, jsonrpc.ClientCloser, error) {
	var options GetStorageMinerOptions
	for _, opt := range opts {
		opt(&options)
	}

	if tn, ok := ctx.App.Metadata["testnode-storage"]; ok {
		return tn.(api.StorageMiner), func() {}, nil
	}

	addr, headers, err := GetRawAPI(ctx, repo.StorageMiner, "v0")
	if err != nil {
		return nil, nil, err
	}

	if options.PreferHttp {
		u, err := url.Parse(addr)
		if err != nil {
			return nil, nil, xerrors.Errorf("parsing miner api URL: %w", err)
		}

		switch u.Scheme {
		case "ws":
			u.Scheme = "http"
		case "wss":
			u.Scheme = "https"
		}

		addr = u.String()
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using miner API v0 endpoint:", addr)
	}

	return client.NewStorageMinerRPCV0(ctx.Context, addr, headers)
}

func GetWorkerAPI(ctx *cli.Context) (api.Worker, jsonrpc.ClientCloser, error) {
	addr, headers, err := GetRawAPI(ctx, repo.Worker, "v0")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using worker API v0 endpoint:", addr)
	}

	return client.NewWorkerRPCV0(ctx.Context, addr, headers)
}

func GetMarketsAPI(ctx *cli.Context) (api.StorageMiner, jsonrpc.ClientCloser, error) {
	// to support lotus-miner cli tests.
	if tn, ok := ctx.App.Metadata["testnode-storage"]; ok {
		return tn.(api.StorageMiner), func() {}, nil
	}

	addr, headers, err := GetRawAPI(ctx, repo.Markets, "v0")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using markets API v0 endpoint:", addr)
	}

	// the markets node is a specialised miner's node, supporting only the
	// markets API, which is a subset of the miner API. All non-markets
	// operations will error out with "unsupported".
	return client.NewStorageMinerRPCV0(ctx.Context, addr, headers)
}

func GetGatewayAPI(ctx *cli.Context) (api.Gateway, jsonrpc.ClientCloser, error) {
	addr, headers, err := GetRawAPI(ctx, repo.FullNode, "v1")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using gateway API v1 endpoint:", addr)
	}

	return client.NewGatewayRPCV1(ctx.Context, addr, headers)
}

func GetGatewayAPIV0(ctx *cli.Context) (v0api.Gateway, jsonrpc.ClientCloser, error) {
	addr, headers, err := GetRawAPI(ctx, repo.FullNode, "v0")
	if err != nil {
		return nil, nil, err
	}

	if IsVeryVerbose {
		_, _ = fmt.Fprintln(ctx.App.Writer, "using gateway API v0 endpoint:", addr)
	}

	return client.NewGatewayRPCV0(ctx.Context, addr, headers)
}

func DaemonContext(cctx *cli.Context) context.Context {
	if mtCtx, ok := cctx.App.Metadata[metadataTraceContext]; ok {
		return mtCtx.(context.Context)
	}

	return context.Background()
}

// ReqContext returns context for cli execution. Calling it for the first time
// installs SIGTERM handler that will close returned context.
// Not safe for concurrent execution.
func ReqContext(cctx *cli.Context) context.Context {
	tCtx := DaemonContext(cctx)

	ctx, done := context.WithCancel(tCtx)
	sigChan := make(chan os.Signal, 2)
	go func() {
		<-sigChan
		done()
	}()
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	return ctx
}
