package route

import (
	"github.com/evergreen-ci/evergreen/rest/data"
	"github.com/evergreen-ci/gimlet"
	"github.com/mongodb/amboy"
)

const defaultLimit = 100

// AttachHandler attaches the api's request handlers to the given mux router.
// It builds a Connector then attaches each of the main functions for
// the api to the router.
func AttachHandler(app *gimlet.APIApp, queue amboy.Queue, URL string, superUsers []string, githubSecret []byte) {
	sc := &data.DBConnector{}

	sc.SetURL(URL)
	sc.SetSuperUsers(superUsers)

	// Middleware
	superUser := gimlet.NewRestrictAccessToUsers(sc.GetSuperUsers())
	checkUser := gimlet.NewRequireAuthHandler()
	addProject := NewProjectContextMiddleware(sc)

	// Routes
	app.AddRoute("/").Version(2).Get().RouteHandler(makePlaceHolderManger(sc))
	app.AddRoute("/admin").Version(2).Get().RouteHandler(makeLegacyAdminConfig(sc))
	app.AddRoute("/admin/banner").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchAdminBanner(sc))
	app.AddRoute("/admin/banner").Version(2).Post().Wrap(superUser).RouteHandler(makeSetAdminBanner(sc))
	app.AddRoute("/admin/events").Version(2).Get().Wrap(superUser).RouteHandler(makeFetchAdminEvents(sc))
	app.AddRoute("/admin/restart").Version(2).Post().Wrap(superUser).RouteHandler(makeRestartRoute(sc, queue))
	app.AddRoute("/admin/revert").Version(2).Post().Wrap(superUser).RouteHandler(makeRevertRouteManager(sc))
	app.AddRoute("/admin/service_flags").Version(2).Post().Wrap(superUser).RouteHandler(makeSetServiceFlagsRouteManager(sc))
	app.AddRoute("/admin/settings").Version(2).Get().Wrap(superUser).RouteHandler(makeFetchAdminSettings(sc))
	app.AddRoute("/admin/settings").Version(2).Post().Wrap(superUser).RouteHandler(makeSetAdminSettings(sc))
	app.AddRoute("/admin/task_queue").Version(2).Delete().Wrap(superUser).RouteHandler(makeClearTaskQueueHandler(sc))
	app.AddRoute("/alias/{name}").Version(2).Get().RouteHandler(makeFetchAliases(sc))
	app.AddRoute("/builds/{build_id}").Version(2).Get().RouteHandler(makeGetBuildByID(sc))
	app.AddRoute("/builds/{build_id}").Version(2).Patch().Wrap(checkUser).RouteHandler(makeChangeStatusForBuild(sc))
	app.AddRoute("/builds/{build_id}/abort").Version(2).Post().Wrap(checkUser).RouteHandler(makeAbortBuild(sc))
	app.AddRoute("/builds/{build_id}/restart").Version(2).Post().Wrap(checkUser).RouteHandler(makeRestartBuild(sc))
	app.AddRoute("/builds/{build_id}/tasks").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchTasksByBuild(sc))
	app.AddRoute("/cost/distro/{distro_id}").Version(2).Get().Wrap(checkUser).RouteHandler(makeCostByDistroHandler(sc))
	app.AddRoute("/cost/project/{project_id}/tasks").Version(2).Get().Wrap(checkUser).RouteHandler(makeTaskCostByProjectRoute(sc))
	app.AddRoute("/cost/version/{version_id}").Version(2).Get().Wrap(checkUser).RouteHandler(makeCostByVersionHandler(sc))
	app.AddRoute("/distros").Version(2).Get().Wrap(checkUser).RouteHandler(makeDistroRoute(sc))
	app.AddRoute("/hooks/github").Version(2).Post().RouteHandler(makeGithubHooksRoute(sc, queue, githubSecret))
	app.AddRoute("/hosts").Version(2).Get().RouteHandler(makeFetchHosts(sc))
	app.AddRoute("/hosts").Version(2).Post().Wrap(checkUser).RouteHandler(makeSpawnHostCreateRoute(sc))
	app.AddRoute("/hosts/{host_id}").Version(2).Get().RouteHandler(makeGetHostByID(sc))
	app.AddRoute("/hosts/{host_id}/change_password").Version(2).Post().Wrap(checkUser).RouteHandler(makeHostChangePassword(sc))
	app.AddRoute("/hosts/{host_id}/extend_expiration").Version(2).Post().Wrap(checkUser).RouteHandler(makeExtendHostExpiration(sc))
	app.AddRoute("/hosts/{host_id}/terminate").Version(2).Post().Wrap(checkUser).RouteHandler(makeTerminateHostRoute(sc))
	app.AddRoute("/hosts/{task_id}/create").Version(2).Post().RouteHandler(makeHostCreateRouteManager(sc))
	app.AddRoute("/hosts/{task_id}/list").Version(2).Get().RouteHandler(makeHostListRouteManager(sc))
	app.AddRoute("/keys").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchKeys(sc))
	app.AddRoute("/keys").Version(2).Post().Wrap(checkUser).RouteHandler(makeSetKey(sc))
	app.AddRoute("/keys/{key_name}").Version(2).Delete().Wrap(checkUser).RouteHandler(makeDeleteKeys(sc))
	app.AddRoute("/patches/{patch_id}").Version(2).Get().RouteHandler(makeFetchPatchByID(sc))
	app.AddRoute("/patches/{patch_id}").Version(2).Patch().Wrap(checkUser).RouteHandler(makeChangePatchStatus(sc))
	app.AddRoute("/patches/{patch_id}/abort").Version(2).Post().Wrap(checkUser).RouteHandler(makeAbortPatch(sc))
	app.AddRoute("/patches/{patch_id}/restart").Version(2).Post().Wrap(checkUser).RouteHandler(makeRestartPatch(sc))
	app.AddRoute("/projects").Version(2).Get().RouteHandler(makeFetchProjectsRoute(sc))
	app.AddRoute("/projects/{project_id}/patches").Version(2).Get().Wrap(checkUser).RouteHandler(makePatchesByProjectRoute(sc))
	app.AddRoute("/projects/{project_id}/versions/tasks").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchProjectTasks(sc))
	app.AddRoute("/projects/{project_id}/recent_versions").Version(2).Get().RouteHandler(makeFetchProjectVersions(sc))
	app.AddRoute("/projects/{project_id}/revisions/{commit_hash}/tasks").Version(2).Get().Wrap(checkUser).RouteHandler(makeTasksByProjectAndCommitHandler(sc))
	app.AddRoute("/status/cli_version").Version(2).Get().RouteHandler(makeFetchCLIVersionRoute(sc))
	app.AddRoute("/status/hosts/distros").Version(2).Get().Wrap(checkUser).RouteHandler(makeHostStatusByDistroRoute(sc))
	app.AddRoute("/status/notifications").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchNotifcationStatusRoute(sc))
	app.AddRoute("/status/recent_tasks").Version(2).Get().RouteHandler(makeRecentTaskStatusHandler(sc))
	app.AddRoute("/subscriptions").Version(2).Delete().Wrap(checkUser).RouteHandler(makeDeleteSubscription(sc))
	app.AddRoute("/subscriptions").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchSubscription(sc))
	app.AddRoute("/subscriptions").Version(2).Post().Wrap(checkUser).RouteHandler(makeSetSubscrition(sc))
	app.AddRoute("/tasks/{task_id}").Version(2).Get().Wrap(checkUser).RouteHandler(makeGetTaskRoute(sc))
	app.AddRoute("/tasks/{task_id}").Version(2).Patch().Wrap(checkUser, addProject).RouteHandler(makeModifyTaskRoute(sc))
	app.AddRoute("/tasks/{task_id}/abort").Version(2).Post().Wrap(checkUser).RouteHandler(makeTaskAbortHandler(sc))
	app.AddRoute("/tasks/{task_id}/generate").Version(2).Post().RouteHandler(makeGenerateTasksHandler(sc))
	app.AddRoute("/tasks/{task_id}/metrics/process").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchTaskProcessMetrics(sc))
	app.AddRoute("/tasks/{task_id}/metrics/system").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchTaskSystmMetrics(sc))
	app.AddRoute("/tasks/{task_id}/restart").Version(2).Post().Wrap(addProject, checkUser).RouteHandler(makeTaskRestartHandler(sc))
	app.AddRoute("/tasks/{task_id}/tests").Version(2).Get().Wrap(addProject).RouteHandler(makeFetchTestsForTask(sc))
	app.AddRoute("/user/settings").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchUserConfig())
	app.AddRoute("/user/settings").Version(2).Post().Wrap(checkUser).RouteHandler(makeSetUserConfig(sc))
	app.AddRoute("/users/{user_id}/hosts").Version(2).Get().Wrap(checkUser).RouteHandler(makeFetchHosts(sc))
	app.AddRoute("/users/{user_id}/patches").Version(2).Get().Wrap(checkUser).RouteHandler(makeUserPatchHandler(sc))
	app.AddRoute("/versions/{version_id}").Version(2).Get().RouteHandler(makeGetVersionByID(sc))
	app.AddRoute("/versions/{version_id}/abort").Version(2).Post().Wrap(checkUser).RouteHandler(makeAbortVersion(sc))
	app.AddRoute("/versions/{version_id}/builds").Version(2).Get().RouteHandler(makeGetVersionBuilds(sc))
	app.AddRoute("/versions/{version_id}/restart").Version(2).Post().Wrap(checkUser).RouteHandler(makeRestartVersion(sc))
}
