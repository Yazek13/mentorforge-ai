"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const test = require("node:test");

// Run the real storage and progress scripts with an isolated browser facade.
// No user browser data is read or changed by these tests.
function page(records = {}, lessons = [lesson()]) {
    const values = new Map(Object.entries(records));
    const events = {};
    function element() {
        return {
            textContent: "", dataset: {}, hidden: false, children: [],
            append(...items) { this.children.push(...items); },
            replaceChildren(...items) { this.children = items; },
        };
    }
    const summary = element();
    const list = element();
    const empty = element();
    const data = { textContent: JSON.stringify(lessons) };
    const storage = {
        getItem(key) { return values.get(key) ?? null; },
        setItem(key, value) { values.set(key, value); },
        removeItem() { throw new Error("progress must not delete results"); },
    };
    const context = vm.createContext({
        localStorage: storage,
        window: { addEventListener(name, fn) { events[name] = fn; } },
        document: {
            addEventListener(name, fn) { events[name] = fn; },
            getElementById() { return data; },
            querySelector(selector) {
                return { "[data-verified-summary]": summary, "[data-verified-list]": list, "[data-verified-empty]": empty }[selector];
            },
            createElement: element,
        },
    });
    for (const file of ["storage.js", "progress.js"]) {
        vm.runInContext(fs.readFileSync(path.join(__dirname, "static", file), "utf8"), context, { filename: file });
    }
    events.DOMContentLoaded();
    return { summary, list, empty, values, storage, events, data };
}

function lesson() {
    return { id: 1, title: "Existing lesson", url: "/lessons/1", question_ids: [1, 2], practice_task_id: 1 };
}

function completeRecords() {
    return {
        "mentorforge.lesson.read.1": "true",
        "mentorforge.question.answer.1": JSON.stringify({ answer: "Own answer", submitted: true }),
        "mentorforge.question.answer.2": JSON.stringify({ answer: "Own second answer", submitted: true }),
        "mentorforge.practice.submission.1": JSON.stringify({ task_id: 1, lesson_id: 1, solution: "Own solution", submitted: true }),
    };
}

test("empty storage does not imply completed learning", () => {
    const result = page();
    assert.equal(result.list.children.length, 0);
    assert.equal(result.empty.hidden, false);
    assert.equal(result.summary.dataset.state, "neutral");
});

test("only existing fully completed lessons appear; user answers are never rendered", () => {
    const records = completeRecords();
    records["mentorforge.lesson.read.999"] = "true";
    const result = page(records);
    assert.equal(result.list.children.length, 1);
    const [link, state] = result.list.children[0].children;
    assert.equal(link.href, "/lessons/1");
    assert.equal(link.textContent, "Existing lesson");
    assert.match(state.textContent, /COMPLETE/);
    assert.doesNotMatch(JSON.stringify(result.list), /Own answer|Own solution/);
    assert.deepEqual(Object.fromEntries(result.values), records);
});

test("each learning step is required, including nonempty submitted practice", () => {
    for (const [key, value] of [
        ["mentorforge.lesson.read.1", "false"],
        ["mentorforge.question.answer.1", JSON.stringify({ answer: "draft", submitted: false })],
        ["mentorforge.question.answer.1", JSON.stringify({ answer: "   ", submitted: true })],
        ["mentorforge.practice.submission.1", "broken JSON"],
        ["mentorforge.practice.submission.1", JSON.stringify({ submitted: true })],
        ["mentorforge.practice.submission.1", JSON.stringify({ task_id: 1, lesson_id: 1, solution: "draft", submitted: false })],
        ["mentorforge.practice.submission.1", JSON.stringify({ task_id: 1, lesson_id: 2, solution: "wrong lesson", submitted: true })],
        ["mentorforge.practice.submission.1", JSON.stringify({ task_id: 2, lesson_id: 1, solution: "wrong task", submitted: true })],
    ]) {
        const records = { ...completeRecords(), [key]: value };
        assert.equal(page(records).list.children.length, 0, key + ": " + value);
    }
});

test("legacy submitted answers remain compatible and old keys are preserved", () => {
    const records = completeRecords();
    delete records["mentorforge.question.answer.1"];
    records["mentorforge.answer.1"] = "Legacy plain answer";
    records["mentorforge.answer.submitted.1"] = "true";
    const result = page(records);
    assert.equal(result.list.children.length, 1);
    assert.equal(result.values.get("mentorforge.answer.1"), "Legacy plain answer");
    assert.equal(JSON.parse(result.values.get("mentorforge.question.answer.1")).answer, "Legacy plain answer");
});

test("string answers without submission metadata remain drafts", () => {
    for (const answer of ["Plain string", JSON.stringify("JSON string")]) {
        assert.equal(page({ ...completeRecords(), "mentorforge.question.answer.1": answer }).list.children.length, 0);
    }
});

test("storage events and restored pages refresh completion without duplicates", () => {
    const result = page(completeRecords());
    result.events.pageshow();
    assert.equal(result.list.children.length, 1);
    result.values.delete("mentorforge.lesson.read.1");
    result.events.storage();
    assert.equal(result.list.children.length, 0);
    assert.equal(result.empty.hidden, false);
});

test("unavailable storage clears stale verification and reports an error", () => {
    const result = page(completeRecords());
    result.storage.getItem = () => { throw new Error("storage denied"); };
    result.events.storage();
    assert.equal(result.list.children.length, 0);
    assert.equal(result.summary.dataset.state, "error");
    assert.equal(result.empty.hidden, true);
});

test("invalid metadata fails safely", () => {
    for (const invalid of [{ ...lesson(), url: "javascript:alert(1)" }, { ...lesson(), question_ids: [] }]) {
        assert.equal(page(completeRecords(), [invalid]).list.children.length, 0);
    }
    const result = page();
    result.data.textContent = "invalid JSON";
    result.events.pageshow();
    assert.equal(result.summary.dataset.state, "error");
});
