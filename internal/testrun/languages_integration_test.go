package testrun

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Opt-in, no model or subscription: actual runtimes exercise the same Adapter,
// FixRunner and BreakageRunner used in a review. Ordinary tests need no Docker.
func TestLanguageContainers(t *testing.T) {
	var images map[string]string
	if raw := os.Getenv("PROVEN_TEST_LANGUAGE_IMAGES"); raw == "" {
		t.Skip("set PROVEN_TEST_LANGUAGE_IMAGES to explicitly run real containers")
	} else if err := json.Unmarshal([]byte(raw), &images); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ adapter, source, path, name, base, head, test, neighbour, neighbourText string }{
		{"pytest", "discount.py", "tests/test_discount.py", "test_discount", "def discount(n):\n    return n * 0.9\n", "def discount(n):\n    return n * 0.8\n", "from discount import discount\ndef test_discount():\n    assert discount(100) == 90\n", "tests/test_neighbour.py", "from discount import discount\ndef test_neighbour():\n    assert discount(100) > 0\n"},
		{"node", "discount.mjs", "discount.test.mjs", "discount once", "export const discount = n => n * 0.9;\n", "export const discount = n => n * 0.8;\n", "import test from 'node:test';\nimport assert from 'node:assert/strict';\nimport {discount} from './discount.mjs';\ntest('discount once', () => assert.equal(discount(100),90));\n", "neighbour.test.mjs", "import test from 'node:test';\nimport assert from 'node:assert/strict';\nimport {discount} from './discount.mjs';\ntest('neighbour', () => assert.ok(discount(100)>0));\n"},
		{"vitest", "discount.ts", "discount.test.ts", "discount once", "export const discount = (n: number) => n * 0.9;\n", "export const discount = (n: number) => n * 0.8;\n", "import {test,expect} from 'vitest';\nimport {discount} from './discount';\ntest('discount once', () => expect(discount(100)).toBe(90));\n", "neighbour.test.ts", "import {test,expect} from 'vitest';\nimport {discount} from './discount';\ntest('neighbour', () => expect(discount(100)).toBeGreaterThan(0));\n"},
	}
	for _, tc := range cases {
		t.Run(tc.adapter, func(t *testing.T) {
			if images[tc.adapter] == "" {
				t.Fatal("missing image for " + tc.adapter)
			}
			base, head := t.TempDir(), t.TempDir()
			write := func(root, p, text string) {
				t.Helper()
				full := filepath.Join(root, p)
				if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			for _, root := range []string{base, head} {
				write(root, tc.neighbour, tc.neighbourText)
				write(root, "package.json", `{"type":"module"}`)
				if err := MakeReadable(root); err != nil {
					t.Fatal(err)
				}
			}
			write(base, tc.source, tc.base)
			write(head, tc.source, tc.head)
			cfg := DefaultContainerConfig()
			cfg.Image = images[tc.adapter]
			runner, err := NewContainer(cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			req := Request{Path: tc.path, Name: tc.name, Content: tc.test, BaseDir: base, ChangeDir: head}
			out, err := runner.Run(ctx, req)
			if err != nil || !out.Found || !out.PassesOnBase || !out.FailsOnChange {
				t.Fatalf("differential: %+v %v", out, err)
			}
			patch := []Edit{{Path: tc.source, Old: "0.8", New: "0.9"}}
			fixed, err := runner.RunFix(ctx, FixRequest{Path: tc.path, Name: tc.name, Content: tc.test, BaseDir: base, ChangeDir: head, Edits: patch})
			if err != nil || len(fixed.Runs) != 4 {
				t.Fatalf("fix: %+v %v", fixed, err)
			}
			for i, want := range []string{ResultPassed, ResultFailed, ResultPassed, ResultPassed} {
				if fixed.Runs[i].Result != want {
					t.Fatalf("fix side %d: %+v", i, fixed.Runs[i])
				}
			}
			breakage, err := runner.RunBreakage(ctx, BreakageRequest{ChangeDir: head, Edits: patch})
			if err != nil || len(breakage.Runs) != 1 || breakage.Runs[0].Result != ResultPassed {
				t.Fatalf("breakage: %+v %v", breakage, err)
			}
			write(head, tc.source, tc.base)
			out, err = runner.Run(ctx, req)
			if err != nil || out.FailsOnChange || !out.PassesOnBase {
				t.Fatalf("control: %+v %v", out, err)
			}
			if tc.adapter == "node" {
				nested := req
				nested.Name = "pricing discount once"
				nested.Content = "import {describe,it} from 'node:test';\nimport assert from 'node:assert/strict';\nimport {discount} from './discount.mjs';\ndescribe('pricing',()=>it('discount once',()=>assert.equal(discount(100),90)));\n"
				out, err = runner.Run(ctx, nested)
				if err != nil || !out.PassesOnBase || out.FailsOnChange {
					t.Fatalf("nested Node test: %+v %v", out, err)
				}
			}
			if tc.adapter == "pytest" {
				class := req
				class.Name = "PriceTests::testDiscount"
				class.Content = "import unittest\nfrom discount import discount\nclass PriceTests(unittest.TestCase):\n    def testDiscount(self):\n        self.assertEqual(discount(100),90)\n"
				out, err = runner.Run(ctx, class)
				if err != nil || !out.PassesOnBase || out.FailsOnChange {
					t.Fatalf("unittest class: %+v %v", out, err)
				}
			}
			if tc.adapter == "vitest" {
				hook := req
				hook.Content = tc.test + "\nimport {afterAll} from 'vitest';\nafterAll(()=>{throw new Error('fixture unavailable')});\n"
				out, err = runner.Run(ctx, hook)
				if err != nil || out.PassesOnBase || out.FailsOnChange || !strings.Contains(out.Output, "could not run") {
					t.Fatalf("failed hook became proof: %+v %v", out, err)
				}
			}
			write(head, tc.source, "this is not valid source code !!!")
			out, err = runner.Run(ctx, req)
			if err != nil || out.FailsOnChange || !strings.Contains(out.Output, "could not run") {
				t.Fatalf("syntax failure became proof: %+v %v", out, err)
			}
		})
	}
}
