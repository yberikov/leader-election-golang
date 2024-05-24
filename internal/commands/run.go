package commands

import (
	"context"
	"fmt"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"strings"
	"time"
)

func InitRunCommand(ctx context.Context) (*cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration from flags and environment variables
			zookeeperServers := strings.Split(viper.GetStringSlice("zk-servers")[0], ",")
			leaderTimeout := viper.GetDuration("leader-timeout")
			attempterTimeout := viper.GetDuration("attempter-timeout")
			fileDir := viper.GetString("file-dir")
			storageCapacity := viper.GetInt("storage-capacity")

			configFile := config.Config{
				ZookeeperServers: zookeeperServers,
				LeaderTimeout:    leaderTimeout,
				AttempterTimeout: attempterTimeout,
				FileDir:          fileDir,
				StorageCapacity:  storageCapacity,
			}

			dg := depgraph.New(configFile)
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("error on: getting logger - %w", err)
			}

			logger.Info("args successfully received", slog.String("servers", strings.Join(zookeeperServers, ", ")))

			runner := run.NewLoopRunner(logger, dg)
			if err != nil {
				return fmt.Errorf("error on: getting runner - %w", err)
			}
			firstState, err := dg.GetInitState()
			if err != nil {
				return fmt.Errorf("error on: getting first state - %w", err)
			}
			err = runner.Run(ctx, firstState)
			if err != nil {
				return fmt.Errorf("error on: running states - %w", err)
			}
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringSliceVarP(&cmdArgs.ZookeeperServers, "zk-servers", "s", []string{"zoo1:2181", "zoo2:2181", "zoo3:2181"}, "Set the zookeeper servers.")
	cmd.Flags().DurationVar(&cmdArgs.LeaderTimeout, "leader-timeout", 10*time.Second, "Leader timeout duration")
	cmd.Flags().DurationVar(&cmdArgs.AttempterTimeout, "attempter-timeout", 10*time.Second, "Attempter timeout duration")
	cmd.Flags().StringVar(&cmdArgs.FileDir, "file-dir", "/tmp/election", "Directory where leader writes files")
	cmd.Flags().IntVar(&cmdArgs.StorageCapacity, "storage-capacity", 10, "Maximum number of files in file-dir")

	// Bind flags to viper

	if err := viper.BindPFlag("zk-servers", cmd.Flags().Lookup("zk-servers")); err != nil {
		return nil, err
	}
	if err := viper.BindPFlag("leader-timeout", cmd.Flags().Lookup("leader-timeout")); err != nil {
		return nil, err
	}
	if err := viper.BindPFlag("attempter-timeout", cmd.Flags().Lookup("attempter-timeout")); err != nil {
		return nil, err
	}
	if err := viper.BindPFlag("file-dir", cmd.Flags().Lookup("file-dir")); err != nil {
		return nil, err
	}
	if err := viper.BindPFlag("storage-capacity", cmd.Flags().Lookup("storage-capacity")); err != nil {
		return nil, err
	}
	return cmd, nil
}

func init() {
	// Read in environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ELECTION")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
