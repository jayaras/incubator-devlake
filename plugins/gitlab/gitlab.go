package main // must be main for plugin entry point

import (
	"github.com/merico-dev/lake/migration"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/gitlab/api"
	"github.com/merico-dev/lake/plugins/gitlab/models/migrationscripts"
	"github.com/merico-dev/lake/plugins/gitlab/tasks"
	"github.com/merico-dev/lake/runner"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	_ "net/http/pprof"
)

var _ core.PluginMeta = (*Gitlab)(nil)
var _ core.PluginInit = (*Gitlab)(nil)
var _ core.PluginTask = (*Gitlab)(nil)
var _ core.PluginApi = (*Gitlab)(nil)
var _ core.Migratable = (*Gitlab)(nil)

type Gitlab string

func (plugin Gitlab) Init(config *viper.Viper, logger core.Logger, db *gorm.DB) error {
	return nil
}

func (plugin Gitlab) Description() string {
	return "To collect and enrich data from Gitlab"
}

func (plugin Gitlab) SubTaskMetas() []core.SubTaskMeta {
	return []core.SubTaskMeta{
		tasks.CollectProjectMeta,
		tasks.ExtractProjectMeta,
		tasks.CollectCommitsMeta,
		tasks.ExtractCommitsMeta,
		tasks.CollectTagMeta,
		tasks.ExtractTagMeta,
		tasks.CollectApiMergeRequestsMeta,
		tasks.ExtractApiMergeRequestsMeta,
		tasks.CollectApiMergeRequestsNotesMeta,
		tasks.ExtractApiMergeRequestsNotesMeta,
		tasks.CollectApiMergeRequestsCommitsMeta,
		tasks.ExtractApiMergeRequestsCommitsMeta,
		tasks.CollectApiPipelinesMeta,
		tasks.ExtractApiPipelinesMeta,
		tasks.CollectApiChildrenOnPipelinesMeta,
		tasks.ExtractApiChildrenOnPipelinesMeta,
		tasks.EnrichMergeRequestsMeta,
		tasks.ConvertProjectMeta,
		tasks.ConvertApiMergeRequestsMeta,
		tasks.ConvertApiCommitsMeta,
		tasks.ConvertApiNotesMeta,
		tasks.ConvertMergeRequestCommentMeta,
		tasks.ConvertApiMergeRequestsCommitsMeta,
	}
}

func (plugin Gitlab) PrepareTaskData(taskCtx core.TaskContext, options map[string]interface{}) (interface{}, error) {
	var op tasks.GitlabOptions
	var err error
	err = mapstructure.Decode(options, &op)
	if err != nil {
		return nil, err
	}

	apiClient, err := tasks.NewGitlabApiClient(taskCtx)
	if err != nil {
		return nil, err
	}

	return &tasks.GitlabTaskData{
		Options:   &op,
		ApiClient: apiClient,
	}, nil
}

func (plugin Gitlab) RootPkgPath() string {
	return "github.com/merico-dev/lake/plugins/gitlab"
}

func (plugin Gitlab) MigrationScripts() []migration.Script {
	return []migration.Script{new(migrationscripts.InitSchemas), new(migrationscripts.UpdateSchemas20220510)}
}

func (plugin Gitlab) ApiResources() map[string]map[string]core.ApiResourceHandler {
	return map[string]map[string]core.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"GET": api.ListConnections,
		},
		"connections/:connectionId": {
			"GET":   api.GetConnection,
			"PATCH": api.PatchConnection,
		},
	}
}

// Export a variable named PluginEntry for Framework to search and load
var PluginEntry Gitlab //nolint

// standalone mode for debugging
func main() {
	gitlabCmd := &cobra.Command{Use: "gitlab"}
	projectId := gitlabCmd.Flags().IntP("project-id", "p", 0, "gitlab project id")

	_ = gitlabCmd.MarkFlagRequired("project-id")
	gitlabCmd.Run = func(cmd *cobra.Command, args []string) {
		runner.DirectRun(cmd, args, PluginEntry, map[string]interface{}{
			"projectId": *projectId,
		})
	}
	runner.RunCmd(gitlabCmd)
}
