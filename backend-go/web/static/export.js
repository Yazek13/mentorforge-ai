"use strict";

(function () {
    function initialize() {
        const exportDataElement = document.getElementById("mentorforge-export-data");
        const exportButtons = document.querySelectorAll("[data-export-results]");
        const exportStatus = document.querySelector("[data-export-status]");

        if (!exportDataElement || exportButtons.length === 0 || !exportStatus) {
            return;
        }

        let exportData;
        try {
            exportData = JSON.parse(exportDataElement.textContent);
        } catch (error) {
            setStatus(exportStatus, "Не удалось подготовить данные для выгрузки.", "error");
            disableButtons(exportButtons);
            return;
        }

        for (const button of exportButtons) {
            button.addEventListener("click", function () {
                try {
                    const submissions = collectSubmissions(exportData.lessons);
                    if (submissions.length === 0) {
                        setStatus(exportStatus, emptyMessage(exportData.scope), "neutral");
                        return;
                    }

                    const markdown = buildMarkdown(submissions);
                    downloadMarkdown(markdown, exportData.file_name);
                    setStatus(exportStatus, "Markdown-файл с результатами подготовлен.", "success");
                } catch (error) {
                    setStatus(exportStatus, "Не удалось подготовить файл с результатами.", "error");
                }
            });
        }
    }

    function collectSubmissions(lessons) {
        const storage = window.MentorForgeStorage;
        if (!storage || !Array.isArray(lessons)) {
            return [];
        }

        const submissions = [];
        for (const lesson of lessons) {
            if (!lesson || !Number.isInteger(lesson.id) || !Array.isArray(lesson.questions)) {
                continue;
            }

            const answers = [];
            for (let index = 0; index < lesson.questions.length; index += 1) {
                const question = lesson.questions[index];
                if (!question || !Number.isInteger(question.id)) {
                    continue;
                }

                const record = storage.readQuestionAnswer(question.id);
                if (!record || record.answer.trim() === "") {
                    continue;
                }

                answers.push({
                    number: index + 1,
                    question: question,
                    record: record,
                });
            }

            const practice = lesson.practice && Number.isInteger(lesson.practice.id)
                ? storage.readPracticeSubmission(lesson.practice.id)
                : null;
            const hasPracticeResult = hasMeaningfulPractice(practice);

            // Не создаём пустой отчёт только из признака чтения лекции.
            if (answers.length === 0 && !hasPracticeResult) {
                continue;
            }

            const lectureRead = storage.isLessonRead(lesson.id);
            const questionsComplete = storage.allQuestionsSubmitted(lesson.questions.map(function (question) {
                return question.id;
            }));
            const practiceComplete = Boolean(practice && practice.submitted && hasPracticeResult);

            submissions.push({
                lesson: lesson,
                answers: answers,
                practice: hasPracticeResult ? practice : null,
                lectureRead: lectureRead,
                complete: lectureRead && questionsComplete && practiceComplete,
            });
        }

        return submissions;
    }

    function hasMeaningfulPractice(record) {
        if (!record) {
            return false;
        }

        const textFields = [
            record.solution,
            record.explanation,
            record.execution_result,
            record.difficulties,
            record.self_assessment,
            record.repository_url,
            record.branch,
            record.commit_hash,
        ];
        if (textFields.some(function (value) { return String(value || "").trim() !== ""; })) {
            return true;
        }

        return record.checks && Object.values(record.checks).some(Boolean);
    }

    function buildMarkdown(submissions) {
        const lines = [
            "# MentorForge AI — результаты обучения",
            "",
            "Дата выгрузки: " + formatDate(new Date()),
            "",
        ];

        for (let lessonIndex = 0; lessonIndex < submissions.length; lessonIndex += 1) {
            const submission = submissions[lessonIndex];
            appendLesson(lines, submission);
            if (lessonIndex + 1 < submissions.length) {
                lines.push("---", "");
            }
        }

        return lines.join("\n");
    }

    function appendLesson(lines, submission) {
        const lesson = submission.lesson;
        lines.push(
            "## Урок: " + oneLine(lesson.title),
            "",
            "* Тема: " + oneLine(lesson.topic),
            "* Уровень: " + oneLine(lesson.level),
            "* Лекция прочитана: " + yesNo(submission.lectureRead),
            "* Урок завершён: " + yesNo(submission.complete),
            "",
            "### Ответы на вопросы",
            ""
        );

        if (submission.answers.length === 0) {
            lines.push("Заполненных ответов пока нет.", "");
        } else {
            for (const answer of submission.answers) {
                lines.push(
                    "#### Вопрос " + answer.number + " (ID " + answer.question.id + "). " + oneLine(answer.question.question),
                    "",
                    "**Мой ответ:**",
                    "",
                    answer.record.answer.trim(),
                    "",
                    "**Ответ зафиксирован:** " + yesNo(answer.record.submitted),
                    "",
                    "**Последнее изменение:** " + formatStoredDate(answer.record.updated_at),
                    ""
                );
            }
        }

        appendPractice(lines, lesson.practice, submission.practice);
    }

    function appendPractice(lines, task, record) {
        lines.push("### Практическое задание", "", "#### Условие", "", oneLine(task.description), "");

        if (Array.isArray(task.requirements) && task.requirements.length > 0) {
            lines.push("**Требования:**", "");
            for (const requirement of task.requirements) {
                lines.push("* " + oneLine(requirement));
            }
            lines.push("");
        }

        if (!record) {
            lines.push("Практическое решение пока не сохранено.", "");
            return;
        }

        lines.push("**Практика завершена:** " + yesNo(record.submitted), "", "**Последнее изменение:** " + formatStoredDate(record.updated_at), "");

        appendFencedSection(lines, "#### Моё решение", record.solution, task.language || "text");
        appendTextSection(lines, "#### Как работает решение", record.explanation);
        appendFencedSection(lines, "#### Результат запуска", record.execution_result, "text");
        appendTextSection(lines, "#### Что было сложно", record.difficulties);

        lines.push("#### Самооценка", "", assessmentLabel(record.self_assessment), "", "#### Что проверено", "");
        appendChecks(lines, record.checks, task.language);
        lines.push("");

        if (record.repository_url || record.branch || record.commit_hash) {
            lines.push("#### GitHub", "");
            if (record.repository_url) {
                lines.push("* Репозиторий: " + oneLine(record.repository_url));
            }
            if (record.branch) {
                lines.push("* Ветка: " + oneLine(record.branch));
            }
            if (record.commit_hash) {
                lines.push("* Commit: " + oneLine(record.commit_hash));
            }
            lines.push("");
        }

        if (record.difficulties.trim() !== "") {
            lines.push("#### Вопросы наставнику", "", record.difficulties.trim(), "");
        }
    }

    function appendTextSection(lines, heading, value) {
        lines.push(heading, "", String(value || "").trim() || "Не заполнено.", "");
    }

    function appendFencedSection(lines, heading, value, language) {
        // Сохраняем фактический вывод и код полностью, включая внутренние отступы и пустые строки.
        const content = String(value || "").replace(/\r\n/g, "\n");
        lines.push(heading, "");
        if (content.trim() === "") {
            lines.push("Не заполнено.", "");
            return;
        }

        const fence = markdownFence(content);
        lines.push(fence + safeFenceLanguage(language), content, fence, "");
    }

    function markdownFence(content) {
        let longest = 0;
        const matches = content.match(/`+/g) || [];
        for (const match of matches) {
            longest = Math.max(longest, match.length);
        }
        return "`".repeat(Math.max(3, longest + 1));
    }

    function safeFenceLanguage(value) {
        const language = String(value || "text").toLowerCase().replace(/[^a-z0-9_+-]/g, "");
        return language || "text";
    }

    function appendChecks(lines, checks, language) {
        const values = checks || {};
        const codeTask = language !== "text";
        const items = [
            { key: "runs", label: "Программа запускается", applicable: codeTask },
            { key: "matches_requirements", label: "Результат соответствует условию", applicable: true },
            { key: "understands_code", label: "Понимаю каждую строку или каждый шаг", applicable: true },
            { key: "checked_errors", label: "Проверил ошибочные входные данные", applicable: codeTask },
            { key: "checked_success_and_error", label: "Проверил успешный и ошибочный сценарии", applicable: codeTask },
            { key: "done_independently", label: "Выполнил самостоятельно", applicable: true },
        ];

        for (const item of items) {
            if (item.applicable) {
                lines.push("* [" + (values[item.key] ? "x" : " ") + "] " + item.label);
            }
        }
    }

    function assessmentLabel(value) {
        const labels = {
            not_understood: "Пока не понял.",
            partially_understood: "Понял частично.",
            understood: "Понял.",
            can_explain: "Могу объяснить другому.",
        };
        return labels[value] || "Не указана.";
    }

    function oneLine(value) {
        return String(value || "").replace(/\s+/g, " ").trim();
    }

    function yesNo(value) {
        return value ? "да" : "нет";
    }

    function formatStoredDate(value) {
        if (!value) {
            return "не указана (старый формат)";
        }
        const date = new Date(value);
        return Number.isNaN(date.getTime()) ? "не указана (старый формат)" : formatDate(date);
    }

    function formatDate(date) {
        const datePart = date.toLocaleDateString("ru-RU", {
            day: "numeric",
            month: "long",
            year: "numeric",
        }).replace(/\s?г\.$/, "");
        const timePart = date.toLocaleTimeString("ru-RU", {
            hour: "2-digit",
            minute: "2-digit",
        });
        return datePart + ", " + timePart;
    }

    function emptyMessage(scope) {
        if (scope === "lesson") {
            return "В этом уроке пока нет сохранённых результатов.";
        }
        if (scope === "topic") {
            return "По этой теме пока нет сохранённых результатов.";
        }
        return "У вас пока нет результатов для выгрузки.";
    }

    function setStatus(element, message, state) {
        element.textContent = message;
        element.dataset.state = state;
    }

    function disableButtons(buttons) {
        for (const button of buttons) {
            button.disabled = true;
        }
    }

    function downloadMarkdown(markdown, fileName) {
        const safeFileName = typeof fileName === "string" && fileName.endsWith(".md")
            ? fileName
            : "mentorforge_submissions.md";
        const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" });
        const objectURL = URL.createObjectURL(blob);
        const link = document.createElement("a");

        link.href = objectURL;
        link.download = safeFileName;
        link.hidden = true;
        document.body.appendChild(link);
        link.click();
        link.remove();

        window.setTimeout(function () {
            URL.revokeObjectURL(objectURL);
        }, 0);
    }

    window.MentorForgeExporter = Object.freeze({
        initialize: initialize,
        collectSubmissions: collectSubmissions,
        buildMarkdown: buildMarkdown,
    });

    document.addEventListener("DOMContentLoaded", initialize);
})();
