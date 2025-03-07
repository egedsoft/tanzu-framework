// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/plugin"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/command"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	local       string
	version     string
	forceDelete bool
)

func init() {
	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	pluginCmd.AddCommand(
		listPluginCmd,
		installPluginCmd,
		upgradePluginCmd,
		describePluginCmd,
		deletePluginCmd,
		repoCmd,
		cleanPluginCmd,
		syncPluginCmd,
		discoverySourceCmd,
	)
	listPluginCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	listPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local discovery/distribution source")
	installPluginCmd.Flags().StringVarP(&local, "local", "l", "", "path to local discovery/distribution source")
	installPluginCmd.Flags().StringVarP(&version, "version", "v", cli.VersionLatest, "version of the plugin")
	deletePluginCmd.Flags().BoolVarP(&forceDelete, "yes", "y", false, "delete the plugin without asking for confirmation")

	command.DeprecateCommand(repoCmd, "")
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage CLI plugins",
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
}

var listPluginCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			var availablePlugins []plugin.Discovered
			if local != "" {
				// get absolute local path
				local, err = filepath.Abs(local)
				if err != nil {
					return err
				}
				availablePlugins, err = pluginmanager.AvailablePluginsFromLocalSource(local)
			} else {
				availablePlugins, err = pluginmanager.AvailablePlugins(serverName)
			}
			if err != nil {
				return err
			}

			data := [][]string{}
			for index := range availablePlugins {
				data = append(data, []string{availablePlugins[index].Name, availablePlugins[index].Description, availablePlugins[index].Scope,
					availablePlugins[index].Source, getInstalledElseAvailablePluginVersion(&availablePlugins[index]), availablePlugins[index].Status})
			}

			output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Name", "Description", "Scope", "Discovery", "Version", "Status")
			for _, row := range data {
				vals := make([]interface{}, len(row))
				for i, val := range row {
					vals[i] = val
				}
				output.AddRow(vals...)
			}
			output.Render()

			return nil
		}
		// TODO: cli.ListPlugins is deprecated: Use pluginmanager.AvailablePluginsFromLocalSource or pluginmanager.AvailablePlugins instead
		descriptors, err := cli.ListPlugins()
		if err != nil {
			return err
		}

		repos := getRepositories()
		plugins, err := repos.ListPlugins()

		// Failure to query plugin metadata from remote repositories should not
		// prevent display of information about plugins already installed.
		if err != nil {
			log.Warningf("Unable to query remote plugin repositories : %v", err)
		}

		data := [][]string{}
		for repoName, descs := range plugins {
			for _, plugin := range descs {
				if plugin.Name == cli.CoreName {
					continue
				}

				status := "not installed"
				var currentVersion string

				repo, err := repos.GetRepository(repoName)
				if err != nil {
					return err
				}

				versionSelector := repo.VersionSelector()
				latestVersion := plugin.FindVersion(versionSelector)
				for _, desc := range descriptors {
					if plugin.Name != desc.Name {
						continue
					}
					status = "installed"
					currentVersion = desc.Version
					compared := semver.Compare(latestVersion, desc.Version)
					if compared == 1 {
						status = "upgrade available"
					}
				}
				data = append(data, []string{plugin.Name, latestVersion, plugin.Description, repoName, currentVersion, status})
			}
		}

		// show plugins installed locally but not found in repositories
		// failure to query the remote repository will degenerate the plugin list
		// to this view as well
		for _, desc := range descriptors {
			var exists bool
			for _, d := range data {
				if desc.Name == d[0] {
					exists = true
					break
				}
			}
			if !exists {
				data = append(data, []string{desc.Name, "", desc.Description, "", desc.Version, "installed"})
			}
		}

		// sort plugins based on their names
		sort.SliceStable(data, func(i, j int) bool {
			return strings.ToLower(data[i][0]) < strings.ToLower(data[j][0])
		})

		output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Name", "Latest Version", "Description", "Repository", "Version", "Status")
		for _, row := range data {
			vals := make([]interface{}, len(row))
			for i, val := range row {
				vals[i] = val
			}
			output.AddRow(vals...)
		}
		output.Render()
		return nil
	},
}

var describePluginCmd = &cobra.Command{
	Use:   "describe [name]",
	Short: "Describe a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}
			pd, err := pluginmanager.DescribePlugin(serverName, pluginName)
			if err != nil {
				return err
			}

			b, err := yaml.Marshal(pd)
			if err != nil {
				return errors.Wrap(err, "could not marshal plugin")
			}
			fmt.Println(string(b))

			return nil
		}

		repos := getRepositories()

		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
		if err != nil {
			return err
		}

		b, err := yaml.Marshal(plugin)
		if err != nil {
			return errors.Wrap(err, "could not marshal plugin")
		}
		fmt.Println(string(b))
		return
	},
}

var installPluginCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {

			// Invoke install plugin from local source if local files are provided
			if local != "" {
				// get absolute local path
				local, err = filepath.Abs(local)
				if err != nil {
					return err
				}
				err = pluginmanager.InstallPluginsFromLocalSource(pluginName, version, local, false)
				if err != nil {
					return err
				}
				if pluginName == cli.AllPlugins {
					log.Successf("successfully installed all plugins")
				} else {
					log.Successf("successfully installed '%s' plugin", pluginName)
				}
				return nil
			}

			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			// Invoke plugin sync if install all plugins is mentioned
			if pluginName == cli.AllPlugins {
				err = pluginmanager.SyncPlugins(serverName)
				if err != nil {
					return err
				}
				log.Successf("successfully installed all plugins")
				return nil
			}

			pluginVersion := version
			if pluginVersion == cli.VersionLatest {
				pluginVersion, err = pluginmanager.GetRecommendedVersionOfPlugin(serverName, pluginName)
				if err != nil {
					return err
				}
			}

			err = pluginmanager.InstallPlugin(serverName, pluginName, pluginVersion)
			if err != nil {
				return err
			}
			log.Successf("successfully installed '%s' plugin", pluginName)
			return nil
		}

		repos := getRepositories()

		if pluginName == cli.AllPlugins {
			return cli.InstallAllMulti(repos)
		}
		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
		if err != nil {
			return err
		}
		if version == cli.VersionLatest {
			version = plugin.FindVersion(repo.VersionSelector())
		}
		err = cli.InstallPlugin(pluginName, version, repo)
		if err != nil {
			return err
		}
		log.Successf("successfully installed %s", pluginName)
		return nil
	},
}

var upgradePluginCmd = &cobra.Command{
	Use:   "upgrade [name]",
	Short: "Upgrade a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			pluginVersion, err := pluginmanager.GetRecommendedVersionOfPlugin(serverName, pluginName)
			if err != nil {
				return err
			}

			err = pluginmanager.UpgradePlugin(serverName, pluginName, pluginVersion)
			if err != nil {
				return err
			}
			log.Successf("successfully upgraded plugin '%s' to version '%s'", pluginName, pluginVersion)
			return nil
		}

		repos := getRepositories()
		repo, err := repos.Find(pluginName)
		if err != nil {
			return err
		}

		plugin, err := repo.Describe(pluginName)
		if err != nil {
			return err
		}

		versionSelector := repo.VersionSelector()
		err = cli.UpgradePlugin(pluginName, plugin.FindVersion(versionSelector), repo)
		if err != nil {
			return err
		}
		log.Successf("successfully upgraded plugin %s", pluginName)
		return nil
	},
}

var deletePluginCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a plugin",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) != 1 {
			return fmt.Errorf("must provide plugin name as positional argument")
		}
		pluginName := args[0]

		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}

			deletePluginOptions := pluginmanager.DeletePluginOptions{
				PluginName:  pluginName,
				ServerName:  serverName,
				ForceDelete: forceDelete,
			}

			err = pluginmanager.DeletePlugin(deletePluginOptions)
			if err != nil {
				return err
			}

			log.Successf("successfully deleted plugin '%s'", pluginName)
			return nil
		}

		err = cli.DeletePlugin(pluginName)
		if err != nil {
			return err
		}
		log.Successf("successfully deleted plugin %s", pluginName)
		return nil
	},
}

var cleanPluginCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			err = pluginmanager.Clean()
			if err != nil {
				return err
			}
			log.Success("successfully cleaned up all plugins")
			return nil
		}

		err = cli.Clean()
		if err != nil {
			return err
		}
		log.Success("successfully cleaned up all plugins")
		return nil
	},
}

var syncPluginCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the plugins",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
			serverName := ""
			server, err := config.GetCurrentServer()
			if err == nil && server != nil {
				serverName = server.Name
			}
			err = pluginmanager.SyncPlugins(serverName)
			if err != nil {
				return err
			}
			log.Success("Done")
			return nil
		}
		return errors.Errorf("command is only applicable if `%s` feature is enabled", cliconfig.FeatureContextAwareCLIForPlugins)
	},
}

func getRepositories() *cli.MultiRepo {
	cfg, err := config.GetClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	if local != "" {
		vs := cli.LoadVersionSelector(cfg.ClientOptions.CLI.UnstableVersionSelector)

		opts := []cli.Option{
			cli.WithVersionSelector(vs),
		}

		m := cli.NewMultiRepo()
		n := filepath.Base(local)
		r := cli.NewLocalRepository(n, local, opts...)
		m.AddRepository(r)
		return m
	}

	return cli.NewMultiRepo(cli.LoadRepositories(cfg)...)
}

// getInstalledElseAvailablePluginVersion return installed plugin version if plugin is installed
// if not installed it returns available recommanded plugin version
func getInstalledElseAvailablePluginVersion(p *plugin.Discovered) string {
	installedOrAvailableVersion := p.InstalledVersion
	if installedOrAvailableVersion == "" {
		installedOrAvailableVersion = p.RecommendedVersion
	}
	return installedOrAvailableVersion
}
