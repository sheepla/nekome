package cli_test

import (
	"testing"

	"github.com/arrow2nd/nekome/cli"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func newCmd(n string) *cli.Command {
	return &cli.Command{
		Name: n,
		Run: func(c *cli.Command, f *pflag.FlagSet) error {
			return nil
		},
	}
}

func TestAddCommand(t *testing.T) {
	c := newCmd("root")
	assert.Equal(
		t,
		len(c.GetChildren()),
		0,
		"サブコマンドが無いか（追加前）",
	)

	// 正常
	c.AddCommand(newCmd("test"))
	assert.Equal(
		t,
		len(c.GetChildren()),
		1,
		"正しくサブコマンドが追加されているか",
	)
}

func TestGetChildren(t *testing.T) {
	r := newCmd("root")
	r.AddCommand(newCmd("a"))
	r.AddCommand(newCmd("b"))

	children := r.GetChildren()

	// 正常
	assert.Equal(
		t,
		children[0].Name,
		"a",
		"サブコマンドを取得できるか（a）",
	)

	assert.Equal(
		t,
		children[1].Name,
		"b",
		"サブコマンドを取得できるか（b）",
	)
}

func TestGetChildrenNames(t *testing.T) {
	c := newCmd("c")
	c.Hidden = true

	r := newCmd("root")
	r.AddCommand(newCmd("a"))
	r.AddCommand(newCmd("b"))
	r.AddCommand(c)

	// 正常
	assert.Equal(
		t,
		r.GetChildrenNames(false),
		[]string{"a", "b"},
		"サブコマンド名を取得できるか",
	)

	assert.NotEqual(
		t,
		r.GetChildrenNames(false),
		[]string{"a", "b", "c"},
		"非表示のコマンドが除外されている",
	)
}

func TestGetChildrenNamesAll(t *testing.T) {
	b := newCmd("b")
	b.AddCommand(newCmd("a"))

	d := newCmd("d")
	d.AddCommand(b)
	d.AddCommand(newCmd("c"))

	r := newCmd("root")
	r.AddCommand(d)
	r.AddCommand(newCmd("e"))

	// 正常
	assert.Equal(
		t,
		r.GetChildrenNames(true),
		[]string{"d", "d b", "d b a", "d c", "e"},
		"全てのサブコマンドの組み合わせを取得できるか",
	)
}

func TestExecute(t *testing.T) {
	r := newCmd("root")
	c := newCmd("neko")

	c.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		return nil
	}

	r.AddCommand(c)

	// 正常
	assert.Nil(t, r.Execute([]string{"neko"}), "実行できるか")

	// 異常
	assert.EqualError(
		t,
		r.Execute([]string{}),
		"no argument",
		"引数が無い",
	)

	assert.EqualError(
		t,
		r.Execute([]string{"hoge"}),
		"command not found: hoge",
		"コマンドが存在しない",
	)
}

func TestExecute_Args(t *testing.T) {
	r := newCmd("root")
	c := newCmd("neko")

	c.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		assert.Equal(t, c.Name, "neko", "コマンド名が取得できるか")
		assert.Equal(t, f.Arg(0), "arg", "引数の取得ができるか")
		return nil
	}

	r.AddCommand(c)

	// 正常
	assert.Nil(t, r.Execute([]string{"neko", "arg"}), "実行できるか")
}

func TestExecute_Flag(t *testing.T) {
	c := newCmd("neko")

	c.SetFlag = func(f *pflag.FlagSet) {
		f.BoolP("kawaii", "k", false, "very kawaii flag")
	}

	c.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		kawaii, _ := f.GetBool("kawaii")
		assert.Equal(t, kawaii, true, "フラグがtrue")
		return nil
	}

	r := newCmd("root")
	r.AddCommand(c)

	// 正常
	assert.Nil(
		t,
		r.Execute([]string{"neko", "--kawaii"}),
		"フラグが指定できるか",
	)

	c.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		kawaii, _ := f.GetBool("kawaii")
		assert.Equal(t, kawaii, false, "フラグがfalse")
		return nil
	}

	// 正常
	assert.Nil(
		t,
		r.Execute([]string{"neko"}),
		"フラグが初期化されているか",
	)
}

func TestExecute_Flag_Arg(t *testing.T) {
	c := newCmd("add")

	c.SetFlag = func(f *pflag.FlagSet) {
		f.String("comment", "", "comment")
	}

	c.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		comment, _ := f.GetString("comment")
		assert.Equal(t, comment, "コメント", "フラグの引数が取得できるか")
		return nil
	}

	r := newCmd("root")
	r.AddCommand(c)

	// 正常
	assert.Nil(
		t,
		r.Execute([]string{"add", "--comment", "コメント"}),
		"実行できるか",
	)
}

func TestExecuteSubCommandArg(t *testing.T) {
	b := newCmd("clone")
	b.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		assert.Equal(t, f.Arg(0), "nekome", "サブコマンドの引数が取得できるか")
		return nil
	}

	a := newCmd("repo")
	a.SetFlag = func(f *pflag.FlagSet) {
		f.Bool("test", false, "")
	}
	a.AddCommand(b)

	r := newCmd("root")
	r.AddCommand(a)

	// 正常
	assert.Nil(
		t,
		r.Execute([]string{"repo", "clone", "nekome"}),
		"実行できるか",
	)

	assert.Nil(
		t,
		r.Execute([]string{"repo"}),
		"サブコマンドを持つコマンドを実行できるか",
	)

	a.Run = func(c *cli.Command, f *pflag.FlagSet) error {
		test, _ := f.GetBool("test")
		assert.Equal(t, test, true, "フラグが取得できるか")
		return nil
	}

	assert.Nil(
		t,
		r.Execute([]string{"repo", "--test"}),
		"サブコマンドを持つコマンドのフラグをパースできるか",
	)
}
