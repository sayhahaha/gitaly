package git

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"gitlab.com/gitlab-org/labkit/correlation"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestFlagValidation(t *testing.T) {
	for _, tt := range []struct {
		option Option
		valid  bool
	}{
		// valid Flag inputs
		{option: Flag{Name: "-k"}, valid: true},
		{option: Flag{Name: "-K"}, valid: true},
		{option: Flag{Name: "--asdf"}, valid: true},
		{option: Flag{Name: "--asdf-qwer"}, valid: true},
		{option: Flag{Name: "--asdf=qwerty"}, valid: true},
		{option: Flag{Name: "-D=A"}, valid: true},
		{option: Flag{Name: "-D="}, valid: true},

		// valid ValueFlag inputs
		{option: ValueFlag{"-k", "adsf"}, valid: true},
		{option: ValueFlag{"-k", "--anything"}, valid: true},
		{option: ValueFlag{"-k", ""}, valid: true},

		// valid ConfigPair inputs
		{option: ConfigPair{Key: "a.b.c", Value: "d"}, valid: true},
		{option: ConfigPair{Key: "core.sound", Value: "meow"}, valid: true},
		{option: ConfigPair{Key: "asdf-qwer.1234-5678", Value: ""}, valid: true},
		{option: ConfigPair{Key: "http.https://user@example.com/repo.git.user", Value: "kitty"}, valid: true},

		// invalid Flag inputs
		{option: Flag{Name: "-*"}},          // invalid character
		{option: Flag{Name: "a"}},           // missing dash
		{option: Flag{Name: "[["}},          // suspicious characters
		{option: Flag{Name: "||"}},          // suspicious characters
		{option: Flag{Name: "asdf=qwerty"}}, // missing dash

		// invalid ValueFlag inputs
		{option: ValueFlag{"k", "asdf"}}, // missing dash

		// invalid ConfigPair inputs
		{option: ConfigPair{Key: "", Value: ""}},            // key cannot be empty
		{option: ConfigPair{Key: " ", Value: ""}},           // key cannot be whitespace
		{option: ConfigPair{Key: "asdf", Value: ""}},        // two components required
		{option: ConfigPair{Key: "asdf.", Value: ""}},       // 2nd component must be non-empty
		{option: ConfigPair{Key: "--asdf.asdf", Value: ""}}, // key cannot start with dash
		{option: ConfigPair{Key: "as[[df.asdf", Value: ""}}, // 1st component cannot contain non-alphanumeric
		{option: ConfigPair{Key: "asdf.as]]df", Value: ""}}, // 2nd component cannot contain non-alphanumeric
	} {
		args, err := tt.option.OptionArgs()
		if tt.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err,
				"expected error, but args %v passed validation", args)
			require.True(t, IsInvalidArgErr(err))
		}
	}
}

func TestGlobalOption(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		option   GlobalOption
		valid    bool
		expected []string
	}{
		{
			desc:     "single-letter flag",
			option:   Flag{Name: "-k"},
			valid:    true,
			expected: []string{"-k"},
		},
		{
			desc:     "long option flag",
			option:   Flag{Name: "--asdf"},
			valid:    true,
			expected: []string{"--asdf"},
		},
		{
			desc:     "multiple single-letter flags",
			option:   Flag{Name: "-abc"},
			valid:    true,
			expected: []string{"-abc"},
		},
		{
			desc:     "single-letter option with value",
			option:   Flag{Name: "-a=value"},
			valid:    true,
			expected: []string{"-a=value"},
		},
		{
			desc:     "long option with value",
			option:   Flag{Name: "--asdf=value"},
			valid:    true,
			expected: []string{"--asdf=value"},
		},
		{
			desc:   "flags without dashes are not allowed",
			option: Flag{Name: "foo"},
			valid:  false,
		},
		{
			desc:   "leading spaces are not allowed",
			option: Flag{Name: " -a"},
			valid:  false,
		},

		{
			desc:     "single-letter value flag",
			option:   ValueFlag{Name: "-a", Value: "value"},
			valid:    true,
			expected: []string{"-a", "value"},
		},
		{
			desc:     "long option value flag",
			option:   ValueFlag{Name: "--foobar", Value: "value"},
			valid:    true,
			expected: []string{"--foobar", "value"},
		},
		{
			desc:     "multiple single-letters for value flag",
			option:   ValueFlag{Name: "-abc", Value: "value"},
			valid:    true,
			expected: []string{"-abc", "value"},
		},
		{
			desc:     "value flag with empty value",
			option:   ValueFlag{Name: "--key", Value: ""},
			valid:    true,
			expected: []string{"--key", ""},
		},
		{
			desc:   "value flag without dashes are not allowed",
			option: ValueFlag{Name: "foo", Value: "bar"},
			valid:  false,
		},
		{
			desc:   "value flag with empty key are not allowed",
			option: ValueFlag{Name: "", Value: "bar"},
			valid:  false,
		},

		{
			desc:     "config pair with key and value",
			option:   ConfigPair{Key: "foo.bar", Value: "value"},
			valid:    true,
			expected: []string{"-c", "foo.bar=value"},
		},
		{
			desc:     "config pair with subsection",
			option:   ConfigPair{Key: "foo.bar.baz", Value: "value"},
			valid:    true,
			expected: []string{"-c", "foo.bar.baz=value"},
		},
		{
			desc:     "config pair without value",
			option:   ConfigPair{Key: "foo.bar"},
			valid:    true,
			expected: []string{"-c", "foo.bar="},
		},
		{
			desc:   "config pair with invalid section format",
			option: ConfigPair{Key: "foo", Value: "value"},
			valid:  false,
		},
		{
			desc:   "config pair with leading whitespace",
			option: ConfigPair{Key: " foo.bar", Value: "value"},
			valid:  false,
		},
		{
			desc:   "config pair with disallowed character in key",
			option: ConfigPair{Key: "http.https://weak.example.com.sslVerify", Value: "false"},
			valid:  false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			args, err := tc.option.GlobalArgs()
			if tc.valid {
				require.NoError(t, err)
				require.Equal(t, tc.expected, args)
			} else {
				require.Error(t, err, "expected error, but args %v passed validation", args)
				require.True(t, IsInvalidArgErr(err))
			}
		})
	}
}

func TestWithConfig(t *testing.T) {
	cfg := config.Cfg{
		BinDir: testhelper.TempDir(t),
		Git: config.Git{
			IgnoreGitconfig: true,
		},
	}

	ctx := testhelper.Context(t)

	gitCmdFactory := newCommandFactory(t, cfg, WithSkipHooks())

	for _, tc := range []struct {
		desc           string
		configPairs    []ConfigPair
		expectedValues map[string]string
	}{
		{
			desc:        "no entries",
			configPairs: []ConfigPair{},
		},
		{
			desc: "single entry",
			configPairs: []ConfigPair{
				{Key: "foo.bar", Value: "baz"},
			},
			expectedValues: map[string]string{
				"foo.bar": "baz",
			},
		},
		{
			desc: "multiple entries",
			configPairs: []ConfigPair{
				{Key: "entry.one", Value: "1"},
				{Key: "entry.two", Value: "2"},
				{Key: "entry.three", Value: "3"},
			},
			expectedValues: map[string]string{
				"entry.one":   "1",
				"entry.two":   "2",
				"entry.three": "3",
			},
		},
		{
			desc: "later entries override previous ones",
			configPairs: []ConfigPair{
				{Key: "override.me", Value: "old value"},
				{Key: "unrelated.entry", Value: "unrelated value"},
				{Key: "override.me", Value: "new value"},
			},
			expectedValues: map[string]string{
				"unrelated.entry": "unrelated value",
				"override.me":     "new value",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			option := WithConfig(tc.configPairs...)

			var commandCfg cmdCfg
			require.NoError(t, option(ctx, cfg, gitCmdFactory, &commandCfg))

			for expectedKey, expectedValue := range tc.expectedValues {
				var stdout bytes.Buffer
				configCmd, err := gitCmdFactory.NewWithoutRepo(ctx, SubCmd{
					Name: "config",
					Args: []string{expectedKey},
				}, WithStdout(&stdout), option)
				require.NoError(t, err)
				require.NoError(t, configCmd.Wait())
				require.Equal(t, expectedValue, text.ChompBytes(stdout.Bytes()))
			}
		})
	}
}

func TestExecCommandFactoryGitalyConfigOverrides(t *testing.T) {
	cfg := config.Cfg{
		BinDir: testhelper.TempDir(t),
		Git: config.Git{
			Config: []config.GitConfig{
				{Key: "foo.bar", Value: "from-gitaly-config"},
			},
			IgnoreGitconfig: true,
		},
	}

	ctx := testhelper.Context(t)

	gitCmdFactory := newCommandFactory(t, cfg, WithSkipHooks())

	var stdout bytes.Buffer
	cmd, err := gitCmdFactory.NewWithoutRepo(ctx,
		SubCmd{
			Name: "config",
			Args: []string{"foo.bar"},
		},
		WithStdout(&stdout),
		WithConfig(ConfigPair{Key: "foo.bar", Value: "from-config-option"}),
		WithConfigEnv(ConfigPair{Key: "foo.bar", Value: "from-config-env"}),
	)
	require.NoError(t, err)
	require.NoError(t, cmd.Wait())
	require.Equal(t, "from-gitaly-config\n", stdout.String())
}

func TestWithConfigEnv(t *testing.T) {
	cfg := config.Cfg{
		BinDir: testhelper.TempDir(t),
		Git: config.Git{
			IgnoreGitconfig: true,
		},
	}

	ctx := testhelper.Context(t)

	gitCmdFactory := newCommandFactory(t, cfg, WithSkipHooks())

	for _, tc := range []struct {
		desc           string
		configPairs    []ConfigPair
		expectedEnv    []string
		expectedValues map[string]string
	}{
		{
			desc:        "no entries",
			configPairs: []ConfigPair{},
			expectedEnv: []string{"GIT_CONFIG_COUNT=0"},
		},
		{
			desc: "single entry",
			configPairs: []ConfigPair{
				{Key: "foo.bar", Value: "baz"},
			},
			expectedEnv: []string{
				"GIT_CONFIG_KEY_0=foo.bar",
				"GIT_CONFIG_VALUE_0=baz",
				"GIT_CONFIG_COUNT=1",
			},
			expectedValues: map[string]string{
				"foo.bar": "baz",
			},
		},
		{
			desc: "multiple entries",
			configPairs: []ConfigPair{
				{Key: "entry.one", Value: "1"},
				{Key: "entry.two", Value: "2"},
				{Key: "entry.three", Value: "3"},
			},
			expectedEnv: []string{
				"GIT_CONFIG_KEY_0=entry.one",
				"GIT_CONFIG_VALUE_0=1",
				"GIT_CONFIG_KEY_1=entry.two",
				"GIT_CONFIG_VALUE_1=2",
				"GIT_CONFIG_KEY_2=entry.three",
				"GIT_CONFIG_VALUE_2=3",
				"GIT_CONFIG_COUNT=3",
			},
			expectedValues: map[string]string{
				"entry.one":   "1",
				"entry.two":   "2",
				"entry.three": "3",
			},
		},
		{
			desc: "later entries override previous ones",
			configPairs: []ConfigPair{
				{Key: "override.me", Value: "old value"},
				{Key: "unrelated.entry", Value: "unrelated value"},
				{Key: "override.me", Value: "new value"},
			},
			expectedEnv: []string{
				"GIT_CONFIG_KEY_0=override.me",
				"GIT_CONFIG_VALUE_0=old value",
				"GIT_CONFIG_KEY_1=unrelated.entry",
				"GIT_CONFIG_VALUE_1=unrelated value",
				"GIT_CONFIG_KEY_2=override.me",
				"GIT_CONFIG_VALUE_2=new value",
				"GIT_CONFIG_COUNT=3",
			},
			expectedValues: map[string]string{
				"unrelated.entry": "unrelated value",
				"override.me":     "new value",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			option := WithConfigEnv(tc.configPairs...)

			var commandCfg cmdCfg
			require.NoError(t, option(ctx, cfg, gitCmdFactory, &commandCfg))
			require.EqualValues(t, tc.expectedEnv, commandCfg.env)

			for expectedKey, expectedValue := range tc.expectedValues {
				var stdout bytes.Buffer
				configCmd, err := gitCmdFactory.NewWithoutRepo(ctx, SubCmd{
					Name: "config",
					Args: []string{expectedKey},
				}, WithStdout(&stdout), option)
				require.NoError(t, err)
				require.NoError(t, configCmd.Wait())
				require.Equal(t, expectedValue, text.ChompBytes(stdout.Bytes()))
			}
		})
	}
}

func TestWithInternalFetch(t *testing.T) {
	cfg := config.Cfg{BinDir: testhelper.TempDir(t)}

	gitCmdFactory := newCommandFactory(t, cfg, WithSkipHooks())
	ctx := testhelper.Context(t)

	md := metadata.Pairs("gitaly-servers", base64.StdEncoding.EncodeToString([]byte(`{"default":{"address":"unix:///tmp/sock","token":"hunter1"}}`)))
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = correlation.ContextWithCorrelation(ctx, "correlation-id-1")

	uploadPackRequest := gitalypb.SSHUploadPackRequest{
		Repository: &gitalypb.Repository{
			StorageName: "default",
		},
	}
	uploadPackRequestMarshalled, err := protojson.Marshal(&uploadPackRequest)
	require.NoError(t, err)

	uploadPackRequestWithSidechannel := gitalypb.SSHUploadPackWithSidechannelRequest{
		Repository: &gitalypb.Repository{
			StorageName: "default",
		},
	}
	uploadPackRequestWithSidechannelMarshalled, err := protojson.Marshal(&uploadPackRequestWithSidechannel)
	require.NoError(t, err)

	for _, tc := range []struct {
		desc                string
		createOption        func() CmdOpt
		expectedSidechannel bool
		expectedPayload     []byte
	}{
		{
			desc: "without sidechannel",
			createOption: func() CmdOpt {
				return WithInternalFetch(&uploadPackRequest)
			},
			expectedSidechannel: false,
			expectedPayload:     uploadPackRequestMarshalled,
		},
		{
			desc: "with sidechannel",
			createOption: func() CmdOpt {
				return WithInternalFetchWithSidechannel(&uploadPackRequestWithSidechannel)
			},
			expectedSidechannel: true,
			expectedPayload:     uploadPackRequestWithSidechannelMarshalled,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var commandCfg cmdCfg

			option := tc.createOption()
			require.NoError(t, option(ctx, cfg, gitCmdFactory, &commandCfg))

			require.Subset(t, commandCfg.env, []string{
				fmt.Sprintf("GIT_SSH_COMMAND=%s upload-pack", cfg.BinaryPath("gitaly-ssh")),
				fmt.Sprintf("GITALY_PAYLOAD=%s", tc.expectedPayload),
				"CORRELATION_ID=correlation-id-1",
				"GIT_SSH_VARIANT=simple",
			})

			if tc.expectedSidechannel {
				require.Contains(t, commandCfg.env, "GITALY_USE_SIDECHANNEL=1")
			} else {
				require.NotContains(t, commandCfg.env, "GITALY_USE_SIDECHANNEL=1")
			}
		})
	}
}

func TestConfigPairsToEnvironment(t *testing.T) {
	for _, tc := range []struct {
		desc        string
		configPairs []ConfigPair
		expectedEnv []string
	}{
		{
			desc: "no pairs",
			expectedEnv: []string{
				"GIT_CONFIG_COUNT=0",
			},
		},
		{
			desc: "single pair",
			configPairs: []ConfigPair{
				{Key: "foo.bar", Value: "baz"},
			},
			expectedEnv: []string{
				"GIT_CONFIG_KEY_0=foo.bar",
				"GIT_CONFIG_VALUE_0=baz",
				"GIT_CONFIG_COUNT=1",
			},
		},
		{
			desc: "multiple pairs",
			configPairs: []ConfigPair{
				{Key: "duplicate.key", Value: "first"},
				{Key: "foo.bar", Value: "baz"},
				{Key: "duplicate.key", Value: "second"},
			},
			expectedEnv: []string{
				"GIT_CONFIG_KEY_0=duplicate.key",
				"GIT_CONFIG_VALUE_0=first",
				"GIT_CONFIG_KEY_1=foo.bar",
				"GIT_CONFIG_VALUE_1=baz",
				"GIT_CONFIG_KEY_2=duplicate.key",
				"GIT_CONFIG_VALUE_2=second",
				"GIT_CONFIG_COUNT=3",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.expectedEnv, ConfigPairsToGitEnvironment(tc.configPairs))
		})
	}
}
