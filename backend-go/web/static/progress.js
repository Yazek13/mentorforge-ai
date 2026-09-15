"use strict";

(function () {
    document.addEventListener("DOMContentLoaded", initializeVerifiedLearning);
    window.addEventListener("storage", initializeVerifiedLearning);
    window.addEventListener("pageshow", initializeVerifiedLearning);

    function initializeVerifiedLearning() {
        const dataElement = document.getElementById("mentorforge-progress-lessons");
        const summary = document.querySelector("[data-verified-summary]");
        const list = document.querySelector("[data-verified-list]");
        const empty = document.querySelector("[data-verified-empty]");
        const storage = window.MentorForgeStorage;

        if (!dataElement || !summary || !list || !empty) {
            return;
        }
        list.replaceChildren();
        if (!storage) {
            showUnavailable(summary, empty);
            return;
        }

        let lessons;
        try {
            lessons = JSON.parse(dataElement.textContent);
        } catch (error) {
            showUnavailable(summary, empty);
            return;
        }
        if (!Array.isArray(lessons)) {
            showUnavailable(summary, empty);
            return;
        }

        const completed = [];
        try {
            for (const lesson of lessons) {
                if (!validLesson(lesson)) {
                    continue;
                }

                const practice = storage.readPracticeSubmission(lesson.practice_task_id);
                const lessonComplete = storage.isLessonRead(lesson.id)
                    && storage.allQuestionsSubmitted(lesson.question_ids)
                    && Boolean(practice && practice.submitted
                        && practice.solution.trim() !== ""
                        && practice.task_id === lesson.practice_task_id
                        && practice.lesson_id === lesson.id);
                if (lessonComplete) {
                    completed.push(lesson);
                }
            }
        } catch (error) {
            showUnavailable(summary, empty);
            return;
        }

        summary.textContent = "Подтверждено: " + completed.length + " из " + lessons.length + " существующих уроков.";
        summary.dataset.state = completed.length > 0 ? "success" : "neutral";
        empty.hidden = completed.length > 0;
        list.replaceChildren();

        for (const lesson of completed) {
            const item = document.createElement("li");
            const link = document.createElement("a");
            const state = document.createElement("strong");
            const details = document.createElement("span");

            link.href = lesson.url;
            link.textContent = lesson.title;
            state.textContent = "COMPLETE ✓";
            details.textContent = "Лекция, " + lesson.question_ids.length + " ответов и практика отмечены завершёнными в этом браузере.";

            item.append(link, state, details);
            list.append(item);
        }
    }

    function validLesson(lesson) {
        return lesson
            && Number.isInteger(lesson.id)
            && lesson.id > 0
            && typeof lesson.title === "string"
            && typeof lesson.url === "string"
            && lesson.url === "/lessons/" + lesson.id
            && Array.isArray(lesson.question_ids)
            && lesson.question_ids.length > 0
            && lesson.question_ids.every(function (id) { return Number.isInteger(id) && id > 0; })
            && Number.isInteger(lesson.practice_task_id)
            && lesson.practice_task_id > 0;
    }

    function showUnavailable(summary, empty) {
        summary.textContent = "Не удалось прочитать локальные результаты этого браузера.";
        summary.dataset.state = "error";
        empty.hidden = true;
    }
})();
