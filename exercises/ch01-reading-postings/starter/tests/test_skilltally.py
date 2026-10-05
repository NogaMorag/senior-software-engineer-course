"""Tests for skilltally. No files outside tmp_path, no network."""

from pathlib import Path

import pytest

from skilltally import Profile, Skill, load_profile, load_skills, mentions, render, tally
from skilltally.catalog import parse_skills
from skilltally.cli import main

GO = Skill("Go", "languages", 3, keywords=("golang",), exact=("Go",))
CPP = Skill("C++", "languages", 33, keywords=("c++",))
JAVA = Skill("Java", "languages", 0, keywords=("java",))
K8S = Skill("Kubernetes", "cloud", 20, keywords=("kubernetes", "k8s"))
SQL = Skill("SQL", "data", 10, keywords=("sql",))
DESIGN = Skill("System design", "craft", 29, keywords=("system design",))


@pytest.mark.parametrize(
    ("text", "skill", "expected"),
    [
        ("We build services in Go and gRPC.", GO, True),
        ("Strong golang experience", GO, True),
        ("Ready to go the extra mile", GO, False),  # exact keywords keep their case
        ("Google Cloud", GO, False),  # "Go" touching "o" is not a match
        ("Modern C++17 codebase", CPP, True),  # the keyword ends in "+": no boundary after it
        ("Experience with c++ or Rust", CPP, True),
        ("JavaScript and TypeScript", JAVA, False),  # "java" touching "S" is not a match
        ("Java/Kotlin backend", JAVA, True),
        ("ניסיון ב־Kubernetes", K8S, True),  # Hebrew prefix with a maqaf
        ("ניסיון בKubernetes", K8S, True),  # Hebrew prefix without one
        ("running on k8s", K8S, True),
        ("PostgreSQL and MySQL", SQL, False),
        ("strong SQL skills", SQL, True),
        ("system\n  design interviews", DESIGN, True),  # spaces match any whitespace
    ],
)
def test_mentions(text: str, skill: Skill, expected: bool) -> None:
    assert mentions(text, skill) is expected


def test_tally_counts_each_posting_once_and_sorts_by_count() -> None:
    postings = {
        "b.txt": "Go, Go, Go. Kubernetes.",
        "a.txt": "Kubernetes and golang",
        "c.txt": "nothing relevant",
    }
    rows = tally(postings, [GO, CPP, K8S])
    assert [(row.skill.name, row.postings) for row in rows] == [
        ("Go", ("a.txt", "b.txt")),
        ("Kubernetes", ("a.txt", "b.txt")),  # a tie keeps the table order
        ("C++", ()),
    ]


def test_render_marks_status_gaps_and_unseen() -> None:
    rows = tally({"1.txt": "Go and k8s", "2.txt": "golang"}, [GO, K8S, CPP])
    text = render(rows, 2, Profile(have=frozenset({"Kubernetes"})), "postings")
    lines = text.splitlines()
    assert lines[0] == "# skilltally: 2 postings in postings"
    assert "| Go | 2/2 | 100% | missing | 3 |" in lines
    assert "| Kubernetes | 1/2 | 50% | have | 20 |" in lines
    assert "Gaps (missing, in at least half the postings): Go" in lines
    assert "Not seen in any posting: C++" in lines


def test_render_without_profile_shows_question_marks() -> None:
    text = render(tally({"1.txt": "Java"}, [JAVA]), 1, None, "p")
    assert "| Java | 1/1 | 100% | ? | - |" in text
    assert "Gaps" not in text


@pytest.mark.parametrize(
    ("data", "message"),
    [
        ({}, "expected at least one"),
        (
            {"skill": [{"name": "Go", "keywords": ["go"]}, {"name": "Go", "keywords": ["x"]}]},
            "twice",
        ),
        ({"skill": [{"name": "Go"}]}, "at least one keyword"),
        ({"skill": [{"name": "Go", "keywords": "go"}]}, "list of non-empty strings"),
        ({"skill": [{"name": "Go", "keywords": ["go"], "chapter": 99}]}, "from 0 to 39"),
    ],
)
def test_parse_skills_rejects_bad_tables(data: dict[str, object], message: str) -> None:
    with pytest.raises(ValueError, match=message):
        parse_skills(data)


def test_load_skills_and_profile_from_files(tmp_path: Path) -> None:
    skills_file = tmp_path / "skills.toml"
    skills_file.write_text(
        '[[skill]]\nname = "Go"\narea = "languages"\nchapter = 3\n'
        'keywords = ["golang"]\nexact = ["Go"]\n',
        encoding="utf-8",
    )
    profile_file = tmp_path / "profile.toml"
    profile_file.write_text('have = ["Rust"]\npartial = ["Go"]\n', encoding="utf-8")
    assert load_skills(skills_file) == [GO]
    profile = load_profile(profile_file)
    assert (profile.status("Rust"), profile.status("Go"), profile.status("Kafka")) == (
        "have",
        "partial",
        "missing",
    )


def write_case(tmp_path: Path) -> tuple[Path, Path]:
    postings = tmp_path / "postings"
    postings.mkdir()
    (postings / "one.txt").write_text("Senior engineer: Go, Kubernetes.", encoding="utf-8")
    (postings / "two.txt").write_text("Backend in golang.", encoding="utf-8")
    (postings / "notes.md").write_text("Kubernetes Kubernetes", encoding="utf-8")  # not a .txt
    skills = tmp_path / "skills.toml"
    skills.write_text(
        '[[skill]]\nname = "Go"\nchapter = 3\nkeywords = ["golang"]\nexact = ["Go"]\n\n'
        '[[skill]]\nname = "Kubernetes"\nchapter = 20\nkeywords = ["kubernetes"]\n',
        encoding="utf-8",
    )
    return postings, skills


def test_cli_prints_the_matrix(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    postings, skills = write_case(tmp_path)
    profile = tmp_path / "profile.toml"
    profile.write_text('have = ["Kubernetes", "Kuberntes"]\n', encoding="utf-8")
    code = main([str(postings), "--skills", str(skills), "--profile", str(profile)])
    out, err = capsys.readouterr()
    assert code == 0
    assert "| Go | 2/2 | 100% | missing | 3 |" in out
    assert "| Kubernetes | 1/2 | 50% | have | 20 |" in out
    assert "unknown skill: 'Kuberntes'" in err  # a typo in the profile is reported, not ignored


def test_cli_without_postings_is_a_usage_error(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    _, skills = write_case(tmp_path)
    assert main([str(tmp_path / "missing"), "--skills", str(skills)]) == 2
    assert "no .txt postings" in capsys.readouterr().err


def test_cli_reports_a_broken_skills_file(
    tmp_path: Path, capsys: pytest.CaptureFixture[str]
) -> None:
    postings, skills = write_case(tmp_path)
    skills.write_text("[[skill]]\nname = \n", encoding="utf-8")
    assert main([str(postings), "--skills", str(skills)]) == 2
    assert capsys.readouterr().err.startswith("skilltally: ")
