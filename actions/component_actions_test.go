package actions

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/madridianfox/elc/core"
	"os"
	"path/filepath"
	"testing"
)

var fakeHomeDir = filepath.Join(os.TempDir(), "home")
var fakeHomeConfigPath = filepath.Join(fakeHomeDir, ".elc.yaml")
var fakeWorkspacePath = filepath.Join(os.TempDir(), "workspaces/project1")

var baseHomeConfig = fmt.Sprintf(`
current_workspace: project1
update_command: update
workspaces:
- name: project1
  path: %s/workspaces/project1
- name: project2
  path: %s/workspaces/project2
`, os.TempDir(), os.TempDir())

func setupMockPc(t *testing.T) *core.MockPC {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPc := core.NewMockPC(ctrl)
	core.Pc = mockPc
	return mockPc
}

func expectReadHomeConfig(mockPC *core.MockPC) {
	mockPC.EXPECT().HomeDir().Return(fakeHomeDir, nil)
	mockPC.EXPECT().FileExists(fakeHomeConfigPath).Return(true)
	mockPC.EXPECT().ReadFile(fakeHomeConfigPath).Return([]byte(baseHomeConfig), nil)
}

func expectReadWorkspaceConfig(mockPC *core.MockPC, workspacePath string, config string, env string) {
	configPath := filepath.Join(workspacePath, "workspace.yaml")
	envPath := filepath.Join(workspacePath, "env.yaml")
	mockPC.EXPECT().Getwd().
		Return(filepath.Join(workspacePath, "apps/test"), nil)
	mockPC.EXPECT().ReadFile(configPath).
		Return([]byte(config), nil)

	envExists := env != ""
	mockPC.EXPECT().FileExists(envPath).
		Return(envExists)
	if envExists {
		mockPC.EXPECT().ReadFile(envPath).
			Return([]byte(env), nil)
	}
}

const workspaceConfig = `
name: ensi
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
`

func TestServiceStart(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfig, "")

	composeFilePath := filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml")

	mockPc.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "", nil)

	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "up", "-d"}, gomock.Any()).
		Return(0, nil)

	_ = StartServiceAction(&core.GlobalOptions{}, []string{})
}

const workspaceConfigWithDeps = `name: ensi
variables:
  USER_ID: "1000"
  GROUP_ID: "1000"
aliases:
  als: dep3
services:
  dep1:
    path: "${WORKSPACE_PATH}/apps/dep1"
  dep2:
    path: "${WORKSPACE_PATH}/apps/dep2"
  dep3:
    path: "${WORKSPACE_PATH}/apps/dep3"
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    dependencies:
      dep1: [default]
      dep2: [default, hook]
      dep3: []
`

func expectStartService(mockPC *core.MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "up", "-d"}, gomock.Any()).
		Return(0, nil)
}

func expectStopService(mockPC *core.MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "asdasd", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "stop"}, gomock.Any()).
		Return(0, nil)
}

func expectDestroyService(mockPC *core.MockPC, composeFilePath string) {
	mockPC.EXPECT().
		ExecToString([]string{"docker", "compose", "-f", composeFilePath, "ps", "--status=running", "-q"}, gomock.Any()).
		Return(0, "asdasd", nil)

	mockPC.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", composeFilePath, "down"}, gomock.Any()).
		Return(0, nil)
}

func TestServiceStartDefaultMode(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StartServiceAction(&core.GlobalOptions{
		Mode: "default",
	}, []string{})
}

func TestServiceStartHookMode(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StartServiceAction(&core.GlobalOptions{
		Mode: "hook",
	}, []string{})
}

func TestServiceStartByName(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = StartServiceAction(&core.GlobalOptions{}, []string{"dep1"})
}

func TestServiceStartByNames(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))

	_ = StartServiceAction(&core.GlobalOptions{}, []string{"dep1", "dep3"})
}

func TestServiceStartByAlias(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))

	_ = StartServiceAction(&core.GlobalOptions{}, []string{"als"})
}

func TestServiceStop(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StopServiceAction(false, []string{}, false, &core.GlobalOptions{})
}

func TestServiceStopByName(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = StopServiceAction(false, []string{"dep1"}, false, &core.GlobalOptions{})
}

func TestServiceStopByNames(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))

	_ = StopServiceAction(false, []string{"dep1", "dep2"}, false, &core.GlobalOptions{})
}

func TestServiceStopAll(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))
	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StopServiceAction(true, []string{}, false, &core.GlobalOptions{})
}

func TestServiceDestroy(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StopServiceAction(false, []string{}, true, &core.GlobalOptions{})
}

func TestServiceDestroyByName(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))

	_ = StopServiceAction(false, []string{"dep1"}, true, &core.GlobalOptions{})
}

func TestServiceDestroyByNames(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))

	_ = StopServiceAction(false, []string{"dep1", "dep2"}, true, &core.GlobalOptions{})
}

func TestServiceDestroyAll(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"))
	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep2/docker-compose.yml"))
	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/dep3/docker-compose.yml"))
	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = StopServiceAction(true, []string{}, true, &core.GlobalOptions{})
}

func TestServiceRestart(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	expectStopService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = RestartServiceAction(false, []string{}, &core.GlobalOptions{})
}

func TestServiceRestartHard(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	expectDestroyService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))

	_ = RestartServiceAction(true, []string{}, &core.GlobalOptions{})
}

func TestServiceCompose(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "some", "command"}, gomock.Any()).
		Return(0, nil)

	_ = ComposeCommandAction(&core.GlobalOptions{}, []string{"some", "command"})
}

func TestServiceComposeByName(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", filepath.Join(fakeWorkspacePath, "apps/dep1/docker-compose.yml"), "some", "command"}, gomock.Any()).
		Return(0, nil)

	_ = ComposeCommandAction(&core.GlobalOptions{
		ComponentName: "dep1",
	}, []string{"some", "command"})
}

func TestServiceExec(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPc.EXPECT().
		IsTerminal().
		Return(true)
	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "1000:1000", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_ = ExecAction(&core.GlobalOptions{
		Cmd: []string{"some", "command"},
		UID: -1,
	})
}

func TestServiceExecWithoutTty(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPc.EXPECT().
		IsTerminal().
		Return(false)
	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "1000:1000", "-T", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_ = ExecAction(&core.GlobalOptions{
		Cmd: []string{"some", "command"},
		UID: -1,
	})
}

func TestServiceExecWithUid(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithDeps, "")

	expectStartService(mockPc, filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"))
	mockPc.EXPECT().
		IsTerminal().
		Return(true)
	mockPc.EXPECT().
		ExecInteractive([]string{"docker", "compose", "-f", filepath.Join(fakeWorkspacePath, "apps/test/docker-compose.yml"), "exec", "-u", "1001", "app", "some", "command"}, gomock.Any()).
		Return(0, nil)

	_ = ExecAction(&core.GlobalOptions{
		Cmd: []string{"some", "command"},
		UID: 1001,
	})
}

const workspaceConfigWithVars = `
name: ensi
variables:
  V_GL: vglobal
  V_GL_SIMPLE_VAR: ${V_GL}-a
  V_GL_WITH_DEFAULT: ${UNDEFINED:-default}
  V_GL_WITH_DEFAULT_VAR: ${UNDEFINED:-$V_GL}
  V_GL_PATHVAR: ./some_folder
  V_GL_PATHVAR_TWO: some_nested_folder
  V_GL_NOT_PATHVAR: ../some_folder
path_variables:
  - V_GL_PATHVAR
  - V_GL_PATHVAR_TWO
services:
  test:
    path: "${WORKSPACE_PATH}/apps/test"
    variables:
      V_IN_SVC: vinsvc
      V_IN_SVC_PATHVAR: ../some_folder_from_service
    path_variables:
      - V_IN_SVC_PATHVAR
  test1:
    path: "${WORKSPACE_PATH}/apps/test1"
    extends: tpl1
    variables:
      V_IN_SVC: vinsvc

templates:
  tpl1:
    path: "${WORKSPACE_PATH}/templates/tpl1"
    variables:
      V_IN_TPL: vintpl
      V_IN_TPL_PATHVAR: ../some_folder_from_service
    path_variables:
      - V_IN_TPL_PATHVAR
`

func TestServiceVars(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithVars, "")

	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("WORKSPACE_PATH=%s/workspaces/project1", os.TempDir())))
	mockPc.EXPECT().Println("WORKSPACE_NAME=ensi")

	mockPc.EXPECT().Println("V_GL=vglobal")
	mockPc.EXPECT().Println("V_GL_SIMPLE_VAR=vglobal-a")
	mockPc.EXPECT().Println("V_GL_WITH_DEFAULT=default")
	mockPc.EXPECT().Println("V_GL_WITH_DEFAULT_VAR=vglobal")
	mockPc.EXPECT().Println(filepath.FromSlash("V_GL_PATHVAR=some_folder"))
	mockPc.EXPECT().Println(filepath.FromSlash("V_GL_PATHVAR_TWO=some_nested_folder"))
	mockPc.EXPECT().Println("V_GL_NOT_PATHVAR=../some_folder")

	mockPc.EXPECT().Println("APP_NAME=test")
	mockPc.EXPECT().Println("COMPOSE_PROJECT_NAME=ensi-test")
	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("SVC_PATH=%s/workspaces/project1/apps/test", os.TempDir())))
	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("COMPOSE_FILE=%s/workspaces/project1/apps/test/docker-compose.yml", os.TempDir())))

	mockPc.EXPECT().Println("V_IN_SVC=vinsvc")
	mockPc.EXPECT().Println(filepath.FromSlash("V_IN_SVC_PATHVAR=../some_folder_from_service"))

	_ = PrintVarsAction(&core.GlobalOptions{}, []string{})
}

func TestServiceVarsWithTpl(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)
	expectReadWorkspaceConfig(mockPc, fakeWorkspacePath, workspaceConfigWithVars, "")

	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("WORKSPACE_PATH=%s/workspaces/project1", os.TempDir())))
	mockPc.EXPECT().Println("WORKSPACE_NAME=ensi")

	mockPc.EXPECT().Println("V_GL=vglobal")
	mockPc.EXPECT().Println("V_GL_SIMPLE_VAR=vglobal-a")
	mockPc.EXPECT().Println("V_GL_WITH_DEFAULT=default")
	mockPc.EXPECT().Println("V_GL_WITH_DEFAULT_VAR=vglobal")
	mockPc.EXPECT().Println(filepath.FromSlash("V_GL_PATHVAR=some_folder"))
	mockPc.EXPECT().Println(filepath.FromSlash("V_GL_PATHVAR_TWO=some_nested_folder"))
	mockPc.EXPECT().Println("V_GL_NOT_PATHVAR=../some_folder")

	mockPc.EXPECT().Println("APP_NAME=test1")
	mockPc.EXPECT().Println("COMPOSE_PROJECT_NAME=ensi-test1")
	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("SVC_PATH=%s/workspaces/project1/apps/test1", os.TempDir())))
	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("TPL_PATH=%s/workspaces/project1/templates/tpl1", os.TempDir())))
	mockPc.EXPECT().Println(filepath.FromSlash(fmt.Sprintf("COMPOSE_FILE=%s/workspaces/project1/templates/tpl1/docker-compose.yml", os.TempDir())))

	mockPc.EXPECT().Println("V_IN_TPL=vintpl")
	mockPc.EXPECT().Println(filepath.FromSlash("V_IN_TPL_PATHVAR=../some_folder_from_service"))
	mockPc.EXPECT().Println("V_IN_SVC=vinsvc")
	mockPc.EXPECT().Println(filepath.FromSlash("V_IN_SVC_PATHVAR=../some_folder_from_service"))

	_ = PrintVarsAction(&core.GlobalOptions{}, []string{"test1"})
}
