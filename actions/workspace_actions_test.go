package actions

import (
	"fmt"
	"github.com/madridianfox/elc/core"
	"os"
	"testing"
)

func TestWorkspaceShow(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	mockPc.EXPECT().Println("project1")

	_ = ShowCurrentWorkspaceAction(&core.GlobalOptions{})
}

func TestWorkspaceList(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	mockPc.EXPECT().Printf("%-10s %s\n", "project1", fmt.Sprintf("%s/workspaces/project1", os.TempDir()))
	mockPc.EXPECT().Printf("%-10s %s\n", "project2", fmt.Sprintf("%s/workspaces/project2", os.TempDir()))

	_ = ListWorkspacesAction()
}

func TestWorkspaceAdd(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	var homeConfigForAdd = fmt.Sprintf(`current_workspace: project1
update_command: update
workspaces:
- name: project1
  path: %s/workspaces/project1
  root_path: ""
- name: project2
  path: %s/workspaces/project2
  root_path: ""
- name: project3
  path: %s/workspaces/project3
  root_path: ""
`, os.TempDir(), os.TempDir(), os.TempDir())

	mockPc.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForAdd), os.FileMode(0644))
	mockPc.EXPECT().Printf("workspace '%s' is added\n", "project3")

	_ = AddWorkspaceAction("project3", fmt.Sprintf("%s/workspaces/project3", os.TempDir()))
}

func TestWorkspaceSelect(t *testing.T) {
	mockPc := setupMockPc(t)
	expectReadHomeConfig(mockPc)

	var homeConfigForSelect = fmt.Sprintf(`current_workspace: project2
update_command: update
workspaces:
- name: project1
  path: %s/workspaces/project1
  root_path: ""
- name: project2
  path: %s/workspaces/project2
  root_path: ""
`, os.TempDir(), os.TempDir())

	mockPc.EXPECT().WriteFile(fakeHomeConfigPath, []byte(homeConfigForSelect), os.FileMode(0644))
	mockPc.EXPECT().Printf("active workspace changed to '%s'\n", "project2")

	_ = SelectWorkspaceAction("project2")
}
