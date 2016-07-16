package git_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/desal/cmd"
	"github.com/desal/dsutil"
	"github.com/desal/git"
	"github.com/desal/richtext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCtx struct {
	bareDir string
	mockDir string
	bareCtx *cmd.Context
	mockCtx *cmd.Context
	gitCtx  *git.Context
}

func SetupMockGit(t *testing.T) *mockCtx {
	m := &mockCtx{}
	var err error
	m.bareDir, err = ioutil.TempDir("", "git_test_bare")
	require.Nil(t, err)

	m.mockDir, err = ioutil.TempDir("", "git_test_mock")
	require.Nil(t, err)

	m.bareCtx = cmd.New(m.bareDir, richtext.Test(t), cmd.Strict, cmd.Warn)
	m.mockCtx = cmd.New(m.mockDir, richtext.Test(t), cmd.Strict, cmd.Warn)

	m.bareCtx.Execf("git --bare init")
	m.mockCtx.Execf("git clone %s .", dsutil.PosixPath(m.bareDir))
	m.mockCtx.Execf("touch init; git add -A; git commit -m init; git push")
	m.gitCtx = git.New(richtext.Test(t))

	return m
}

func (m *mockCtx) Close() {
	os.RemoveAll(m.bareDir)
	os.RemoveAll(m.mockDir)
}

func TestStatus(t *testing.T) {
	m := SetupMockGit(t)
	defer m.Close()

	status, err := m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.Clean, status)

	m.mockCtx.Execf("touch mod")
	status, err = m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.Uncommitted, status)

	m.mockCtx.Execf("git add -A")
	status, err = m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.Uncommitted, status)

	m.mockCtx.Execf("git commit -m mod")
	status, err = m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.NoUpstream, status)

	m.mockCtx.Execf("git push origin master:newbranch")
	status, err = m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.NotMaster, status)

	m.mockCtx.Execf("git push origin master")
	status, err = m.gitCtx.Status(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, git.Clean, status)

}

func TestTopLevel(t *testing.T) {
	m := SetupMockGit(t)
	defer m.Close()

	toplevel, err := m.gitCtx.TopLevel(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, m.mockDir, toplevel)

	m.mockCtx.Execf("mkdir -p a/b/c")
	toplevel, err = m.gitCtx.TopLevel(m.mockDir + "/a/b/c")
	assert.Nil(t, err)
	assert.Equal(t, m.mockDir, toplevel)
}

func TestTags(t *testing.T) {
	m := SetupMockGit(t)
	defer m.Close()

	tags, err := m.gitCtx.Tags(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, []string{}, tags)
	mostRecentTag, err := m.gitCtx.MostRecentTag(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, "", mostRecentTag)

	m.mockCtx.Execf("git tag v1.0.0; git tag Awesome")
	tags, err = m.gitCtx.Tags(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, []string{"Awesome", "v1.0.0"}, tags)
	mostRecentTag, err = m.gitCtx.MostRecentTag(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, "Awesome", mostRecentTag)

	m.mockCtx.Execf("touch mod; git add -A; git commit -m mod")
	tags, err = m.gitCtx.Tags(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, []string{}, tags)
	mostRecentTag, err = m.gitCtx.MostRecentTag(m.mockDir)
	assert.Nil(t, err)
	assert.Equal(t, "Awesome", mostRecentTag)

}
