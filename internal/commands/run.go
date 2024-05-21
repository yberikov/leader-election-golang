package commands

import (
	"fmt"
	config2 "github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InitRunCommand() (*cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, args []string) error {

			zookeeperServers := strings.Split(os.Getenv("ELECTION_ZK_SERVERS"), ",")
			leaderTimeout, _ := time.ParseDuration(os.Getenv("ELECTION_LEADER_TIMEOUT"))
			attempterTimeout, _ := time.ParseDuration(os.Getenv("ELECTION_ATTEMPTER_TIMEOUT"))
			fileDir := os.Getenv("ELECTION_FILE_DIR")
			storageCapacity, _ := strconv.Atoi(os.Getenv("ELECTION_STORAGE_CAPACITY"))

			config := config2.Config{
				ZookeeperServers: zookeeperServers,
				LeaderTimeout:    leaderTimeout,
				AttempterTimeout: attempterTimeout,
				FileDir:          fileDir,
				StorageCapacity:  storageCapacity,
			}

			dg := depgraph.New(config)
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}

			logger.Info("args received", slog.String("servers", strings.Join(cmdArgs.ZookeeperServers, ", ")))

			runner, err := dg.GetRunner()
			if err != nil {
				return fmt.Errorf("get runner: %w", err)
			}
			firstState, err := dg.GetInitState()
			if err != nil {
				return fmt.Errorf("get first state: %w", err)
			}
			err = runner.Run(cmd.Context(), firstState)
			if err != nil {
				return fmt.Errorf("run states: %w", err)
			}
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringSliceVarP(&cmdArgs.ZookeeperServers, "zk-servers", "s", []string{}, "Set the zookeeper servers.")
	cmd.Flags().DurationVar(&cmdArgs.LeaderTimeout, "leader-timeout", 0, "Leader timeout duration")
	cmd.Flags().DurationVar(&cmdArgs.AttempterTimeout, "attempter-timeout", 0, "Attempter timeout duration")
	cmd.Flags().StringVar(&cmdArgs.FileDir, "file-dir", "", "Directory where leader writes files")
	cmd.Flags().IntVar(&cmdArgs.StorageCapacity, "storage-capacity", 0, "Maximum number of files in file-dir")

	return cmd, nil

}

func initConfig() {
	// Read in environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ELECTION")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
