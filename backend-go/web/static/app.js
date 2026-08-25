"use strict";

(function () {
    document.addEventListener("DOMContentLoaded", function () {
        const runtime = readRuntimeData();
        if (!runtime || !window.MentorForgeStorage) {
            return;
        }

        updateStageIndicators(runtime);

        const page = document.querySelector("[data-lesson-page]");
        if (!page) {
            return;
        }

        switch (page.dataset.lessonPage) {
        case "lecture":
            initializeLecture(runtime);
            break;
        case "questions":
            initializeQuestions(runtime);
            break;
        case "practice":
            initializePractice(runtime);
            break;
        case "complete":
            initializeCompletion(runtime);
            break;
        }
    });

    function readRuntimeData() {
        const element = document.getElementById("mentorforge-lesson-data");
        if (!element) {
            return null;
        }

        try {
            const data = JSON.parse(element.textContent);
            if (!Number.isInteger(data.lesson_id) || data.lesson_id <= 0 || !Array.isArray(data.question_ids)) {
                return null;
            }
            return data;
        } catch (error) {
            return null;
        }
    }

    function initializeLecture(runtime) {
        const storage = window.MentorForgeStorage;
        const button = document.querySelector("[data-mark-lesson-read]");
        const status = document.querySelector("[data-lesson-read-status]");
        if (!button || !status) {
            return;
        }

        try {
            if (storage.isLessonRead(runtime.lesson_id)) {
                setStatus(status, "Лекция уже отмечена прочитанной. Можно перейти к вопросам.", "success");
            }
        } catch (error) {
            setStatus(status, "Не удалось прочитать состояние урока в браузере.", "error");
        }

        button.addEventListener("click", function () {
            try {
                storage.markLessonRead(runtime.lesson_id);
                updateStageIndicators(runtime);
                const nextURL = button.dataset.nextUrl;
                if (nextURL) {
                    window.location.assign(nextURL);
                }
            } catch (error) {
                setStatus(status, "Не удалось сохранить отметку о чтении в этом браузере.", "error");
            }
        });
    }

    function initializeQuestions(runtime) {
        const storage = window.MentorForgeStorage;
        const lock = document.querySelector('[data-stage-lock="questions"]');
        const workflow = document.querySelector("[data-questions-workflow]");
        if (!lock || !workflow) {
            return;
        }

        try {
            if (!storage.isLessonRead(runtime.lesson_id)) {
                return;
            }
        } catch (error) {
            return;
        }

        lock.hidden = true;
        workflow.hidden = false;

        const editors = workflow.querySelectorAll("[data-question-editor]");
        for (const editor of editors) {
            initializeQuestionEditor(editor, runtime);
        }
        updateQuestionProgress(runtime);
    }

    function initializeQuestionEditor(editor, runtime) {
        const storage = window.MentorForgeStorage;
        const questionID = Number(editor.dataset.questionId);
        const textarea = editor.querySelector("[data-answer-text]");
        const saveButton = editor.querySelector("[data-save-answer]");
        const submitButton = editor.querySelector("[data-submit-answer]");
        const editButton = editor.querySelector("[data-edit-answer]");
        const status = editor.querySelector("[data-answer-status]");
        const comparison = editor.querySelector("[data-answer-comparison]");
        const ownAnswer = editor.querySelector("[data-own-answer-display]");

        if (!Number.isInteger(questionID) || questionID <= 0 || !textarea || !saveButton || !submitButton || !editButton || !status || !comparison || !ownAnswer) {
            return;
        }

        try {
            const saved = storage.readQuestionAnswer(questionID);
            if (saved && saved.answer.trim() !== "") {
                applyRecord(saved);
            }
        } catch (error) {
            setStatus(status, "Не удалось прочитать сохранённый ответ.", "error");
        }

        textarea.addEventListener("input", function () {
            setStatus(status, "Есть несохранённые изменения.", "neutral");
        });

        saveButton.addEventListener("click", function () {
            save(false);
        });

        submitButton.addEventListener("click", function () {
            save(true);
        });

        editButton.addEventListener("click", function () {
            try {
                const record = storage.saveQuestionAnswer(questionID, textarea.value, false);
                applyRecord(record);
                textarea.focus();
                updateQuestionProgress(runtime);
            } catch (error) {
                setStatus(status, "Не удалось открыть ответ для изменения.", "error");
            }
        });

        function save(submitted) {
            const answer = textarea.value.trim();
            if (answer === "") {
                setStatus(status, "Сначала напишите ответ своими словами.", "error");
                return;
            }

            try {
                const record = storage.saveQuestionAnswer(questionID, answer, submitted);
                applyRecord(record);
                updateQuestionProgress(runtime);
            } catch (error) {
                setStatus(status, "Не удалось сохранить ответ в этом браузере.", "error");
            }
        }

        function applyRecord(record) {
            textarea.value = record.answer;
            textarea.readOnly = record.submitted;
            saveButton.hidden = record.submitted;
            submitButton.hidden = record.submitted;
            editButton.hidden = !record.submitted;
            comparison.hidden = !record.submitted;
            ownAnswer.textContent = record.answer;

            const date = formatStoredDate(record.updated_at);
            if (record.submitted) {
                setStatus(status, "Ответ зафиксирован · " + date, "success");
            } else {
                setStatus(status, "Черновик сохранён · " + date, "success");
            }
        }
    }

    function updateQuestionProgress(runtime) {
        const storage = window.MentorForgeStorage;
        const progress = document.querySelector("[data-question-progress]");
        const disabledButton = document.querySelector("[data-practice-disabled]");
        const practiceLink = document.querySelector("[data-practice-link]");
        const submittedCount = storage.submittedQuestionCount(runtime.question_ids);
        const allSubmitted = storage.allQuestionsSubmitted(runtime.question_ids);

        if (progress) {
            progress.textContent = "Зафиксировано " + submittedCount + " из " + runtime.question_ids.length + ".";
        }
        if (disabledButton) {
            disabledButton.hidden = allSubmitted;
        }
        if (practiceLink) {
            practiceLink.hidden = !allSubmitted;
        }
        updateStageIndicators(runtime);
    }

    function initializePractice(runtime) {
        const storage = window.MentorForgeStorage;
        const lock = document.querySelector('[data-stage-lock="practice"]');
        const workflow = document.querySelector("[data-practice-workflow]");
        const lockMessage = document.querySelector("[data-practice-lock-message]");
        if (!lock || !workflow) {
            return;
        }

        let lectureRead = false;
        let questionsComplete = false;
        try {
            lectureRead = storage.isLessonRead(runtime.lesson_id);
            questionsComplete = storage.allQuestionsSubmitted(runtime.question_ids);
        } catch (error) {
            if (lockMessage) {
                lockMessage.textContent = "Не удалось проверить локальное состояние урока.";
            }
            return;
        }

        if (!lectureRead || !questionsComplete) {
            if (lockMessage) {
                lockMessage.textContent = !lectureRead
                    ? "Сначала прочитайте лекцию, затем зафиксируйте ответы на все вопросы."
                    : "Зафиксируйте ответы на все обязательные вопросы урока.";
            }
            return;
        }

        lock.hidden = true;
        workflow.hidden = false;
        initializePracticeForm(runtime);
    }

    function initializePracticeForm(runtime) {
        const storage = window.MentorForgeStorage;
        const form = document.querySelector("[data-practice-form]");
        if (!form) {
            return;
        }

        const taskID = Number(form.dataset.taskId);
        const lessonID = Number(form.dataset.lessonId);
        const status = form.querySelector("[data-practice-status]");
        const saveButton = form.querySelector("[data-save-practice]");
        const submitButton = form.querySelector("[data-submit-practice]");
        const editButton = form.querySelector("[data-edit-practice]");
        const clearButton = form.querySelector("[data-clear-practice]");
        const completeLink = form.querySelector("[data-complete-link]");
        if (!Number.isInteger(taskID) || taskID <= 0 || !Number.isInteger(lessonID) || lessonID <= 0 || !status || !saveButton || !submitButton || !editButton || !clearButton || !completeLink) {
            return;
        }

        form.addEventListener("submit", function (event) {
            event.preventDefault();
        });

        try {
            const saved = storage.readPracticeSubmission(taskID);
            if (saved) {
                applyPracticeRecord(form, saved, status, saveButton, submitButton, editButton, completeLink);
            }
        } catch (error) {
            setStatus(status, "Не удалось восстановить локально сохранённое решение.", "error");
        }

        for (const field of form.querySelectorAll("[data-practice-field], [data-practice-check]")) {
            field.addEventListener("input", function () {
                if (!field.disabled) {
                    setStatus(status, "Есть несохранённые изменения.", "neutral");
                }
            });
        }

        saveButton.addEventListener("click", function () {
            save(false);
        });

        submitButton.addEventListener("click", function () {
            const solution = readPracticeField(form, "solution").trim();
            const assessment = readPracticeField(form, "self_assessment");
            if (solution === "") {
                setStatus(status, "Добавьте код или текст собственного решения.", "error");
                return;
            }
            if (assessment === "") {
                setStatus(status, "Выберите самооценку перед завершением практики.", "error");
                return;
            }
            save(true);
        });

        editButton.addEventListener("click", function () {
            try {
                const record = storage.savePracticeSubmission(collectPracticeRecord(form, taskID, lessonID, false));
                applyPracticeRecord(form, record, status, saveButton, submitButton, editButton, completeLink);
                const solution = form.querySelector('[data-practice-field="solution"]');
                if (solution) {
                    solution.focus();
                }
                updateStageIndicators(runtime);
            } catch (error) {
                setStatus(status, "Не удалось открыть решение для изменения.", "error");
            }
        });

        clearButton.addEventListener("click", function () {
            const confirmed = window.confirm("Очистить только решение этого практического задания? Ответы других уроков останутся без изменений.");
            if (!confirmed) {
                return;
            }

            try {
                storage.clearPracticeSubmission(taskID);
                clearPracticeForm(form);
                setPracticeFormLocked(form, false);
                saveButton.hidden = false;
                submitButton.hidden = false;
                editButton.hidden = true;
                completeLink.hidden = true;
                setStatus(status, "Решение очищено.", "success");
                updateStageIndicators(runtime);
            } catch (error) {
                setStatus(status, "Не удалось очистить решение.", "error");
            }
        });

        function save(submitted) {
            try {
                const record = storage.savePracticeSubmission(collectPracticeRecord(form, taskID, lessonID, submitted));
                applyPracticeRecord(form, record, status, saveButton, submitButton, editButton, completeLink);
                updateStageIndicators(runtime);
            } catch (error) {
                setStatus(status, "Не удалось сохранить решение в этом браузере.", "error");
            }
        }
    }

    function collectPracticeRecord(form, taskID, lessonID, submitted) {
        const checks = {
            runs: false,
            matches_requirements: false,
            understands_code: false,
            checked_errors: false,
            checked_success_and_error: false,
            done_independently: false,
        };
        for (const checkbox of form.querySelectorAll("[data-practice-check]")) {
            checks[checkbox.dataset.practiceCheck] = checkbox.checked;
        }

        return {
            task_id: taskID,
            lesson_id: lessonID,
            solution: readPracticeField(form, "solution"),
            explanation: readPracticeField(form, "explanation"),
            execution_result: readPracticeField(form, "execution_result"),
            difficulties: readPracticeField(form, "difficulties"),
            checks: checks,
            self_assessment: readPracticeField(form, "self_assessment"),
            repository_url: readPracticeField(form, "repository_url"),
            branch: readPracticeField(form, "branch"),
            commit_hash: readPracticeField(form, "commit_hash"),
            updated_at: "",
            submitted: submitted,
        };
    }

    function readPracticeField(form, name) {
        const field = form.querySelector('[data-practice-field="' + name + '"]');
        return field ? field.value : "";
    }

    function applyPracticeRecord(form, record, status, saveButton, submitButton, editButton, completeLink) {
        for (const field of form.querySelectorAll("[data-practice-field]")) {
            field.value = record[field.dataset.practiceField] || "";
        }
        for (const checkbox of form.querySelectorAll("[data-practice-check]")) {
            checkbox.checked = Boolean(record.checks && record.checks[checkbox.dataset.practiceCheck]);
        }

        setPracticeFormLocked(form, record.submitted);
        saveButton.hidden = record.submitted;
        submitButton.hidden = record.submitted;
        editButton.hidden = !record.submitted;
        completeLink.hidden = !record.submitted;

        const date = formatStoredDate(record.updated_at);
        setStatus(
            status,
            record.submitted ? "Практика завершена · " + date : "Черновик сохранён · " + date,
            "success"
        );
    }

    function setPracticeFormLocked(form, locked) {
        for (const field of form.querySelectorAll("[data-practice-field], [data-practice-check]")) {
            field.disabled = locked;
        }
    }

    function clearPracticeForm(form) {
        for (const field of form.querySelectorAll("[data-practice-field]")) {
            field.value = "";
        }
        for (const checkbox of form.querySelectorAll("[data-practice-check]")) {
            checkbox.checked = false;
        }
    }

    function initializeCompletion(runtime) {
        const storage = window.MentorForgeStorage;
        const message = document.querySelector("[data-completion-message]");
        const lectureText = document.querySelector("[data-completion-lecture]");
        const questionsText = document.querySelector("[data-completion-questions]");
        const practiceText = document.querySelector("[data-completion-practice]");

        try {
            const lectureRead = storage.isLessonRead(runtime.lesson_id);
            const submittedCount = storage.submittedQuestionCount(runtime.question_ids);
            const questionsComplete = storage.allQuestionsSubmitted(runtime.question_ids);
            const practice = storage.readPracticeSubmission(runtime.practice_task_id);
            const practiceComplete = Boolean(practice && practice.submitted);
            const complete = lectureRead && questionsComplete && practiceComplete;

            if (lectureText) {
                lectureText.textContent = lectureRead ? "Прочитана" : "Не прочитана";
            }
            if (questionsText) {
                questionsText.textContent = submittedCount + " из " + runtime.question_ids.length + " зафиксировано";
            }
            if (practiceText) {
                practiceText.textContent = practiceComplete ? "Завершена" : "Не завершена";
            }
            if (message) {
                message.textContent = complete
                    ? "Урок завершён. Можно скачать результаты и перейти дальше."
                    : "Урок ещё не завершён. Вернитесь к отмеченным этапам и закончите их.";
                message.dataset.state = complete ? "success" : "neutral";
            }

            updateCompletionCard("lecture", lectureRead);
            updateCompletionCard("questions", questionsComplete);
            updateCompletionCard("practice", practiceComplete);

            const nextDisabled = document.querySelector("[data-next-lesson-disabled]");
            const nextLink = document.querySelector("[data-next-lesson]");
            if (nextDisabled) {
                nextDisabled.hidden = complete;
            }
            if (nextLink) {
                nextLink.hidden = !complete;
            }
        } catch (error) {
            if (message) {
                message.textContent = "Не удалось прочитать локальные результаты урока.";
                message.dataset.state = "error";
            }
        }

        const resetButton = document.querySelector("[data-reset-lesson]");
        if (resetButton) {
            resetButton.addEventListener("click", function () {
                const confirmed = window.confirm("Начать урок заново и удалить его локальные ответы и практическое решение? Данные других уроков останутся без изменений.");
                if (!confirmed) {
                    return;
                }

                try {
                    storage.clearLessonRead(runtime.lesson_id);
                    for (const questionID of runtime.question_ids) {
                        storage.clearQuestionAnswer(Number(questionID));
                    }
                    storage.clearPracticeSubmission(runtime.practice_task_id);
                    window.location.assign("/lessons/" + runtime.lesson_id);
                } catch (error) {
                    if (message) {
                        message.textContent = "Не удалось очистить локальные данные урока.";
                        message.dataset.state = "error";
                    }
                }
            });
        }
    }

    function updateCompletionCard(name, complete) {
        const card = document.querySelector('[data-completion-card="' + name + '"]');
        if (card) {
            card.classList.toggle("is-complete", complete);
        }
    }

    function updateStageIndicators(runtime) {
        const storage = window.MentorForgeStorage;
        const stages = document.querySelector("[data-lesson-stages]");
        if (!storage || !stages) {
            return;
        }

        try {
            const lectureRead = storage.isLessonRead(runtime.lesson_id);
            const questionsComplete = storage.allQuestionsSubmitted(runtime.question_ids);
            const practice = storage.readPracticeSubmission(runtime.practice_task_id);
            const practiceComplete = Boolean(practice && practice.submitted);
            const lessonComplete = lectureRead && questionsComplete && practiceComplete;

            toggleStage(stages, "lecture", lectureRead);
            toggleStage(stages, "questions", questionsComplete);
            toggleStage(stages, "practice", practiceComplete);
            toggleStage(stages, "complete", lessonComplete);
        } catch (error) {
            stages.dataset.state = "unavailable";
        }
    }

    function toggleStage(stages, name, complete) {
        const item = stages.querySelector('[data-stage-item="' + name + '"]');
        if (item) {
            item.classList.toggle("is-complete", complete);
        }
    }

    function setStatus(element, message, state) {
        element.textContent = message;
        element.dataset.state = state;
    }

    function formatStoredDate(value) {
        if (!value) {
            return "дата не указана";
        }
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) {
            return "дата не указана";
        }
        return date.toLocaleString("ru-RU", {
            day: "numeric",
            month: "long",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }
})();
