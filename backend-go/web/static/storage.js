"use strict";

(function () {
    const lessonReadPrefix = "mentorforge.lesson.read.";
    const questionAnswerPrefix = "mentorforge.question.answer.";
    const practiceSubmissionPrefix = "mentorforge.practice.submission.";

    // Эти ключи использовались первой версией интерфейса и читаются для миграции.
    const legacyAnswerPrefix = "mentorforge.answer.";
    const legacyUpdatedAtPrefix = "mentorforge.answer.updated_at.";
    const legacySubmittedPrefix = "mentorforge.answer.submitted.";

    function isPositiveInteger(value) {
        return Number.isInteger(value) && value > 0;
    }

    function normalizeBoolean(value) {
        return value === true || value === 1 || value === "1" || value === "true";
    }

    function normalizeString(value) {
        return typeof value === "string" ? value : "";
    }

    function createQuestionRecord(answer, updatedAt, submitted) {
        return {
            answer: normalizeString(answer),
            updated_at: normalizeString(updatedAt),
            submitted: normalizeBoolean(submitted),
        };
    }

    function parseQuestionRecord(rawValue, legacyUpdatedAt, legacySubmitted) {
        if (rawValue === null) {
            return null;
        }

        try {
            const parsed = JSON.parse(rawValue);
            if (typeof parsed === "string") {
                return createQuestionRecord(parsed, legacyUpdatedAt, legacySubmitted);
            }
            if (parsed && typeof parsed === "object" && !Array.isArray(parsed) && typeof parsed.answer === "string") {
                return createQuestionRecord(
                    parsed.answer,
                    typeof parsed.updated_at === "string" ? parsed.updated_at : legacyUpdatedAt,
                    parsed.submitted === undefined ? legacySubmitted : parsed.submitted
                );
            }
        } catch (error) {
            return createQuestionRecord(rawValue, legacyUpdatedAt, legacySubmitted);
        }

        return createQuestionRecord(rawValue, legacyUpdatedAt, legacySubmitted);
    }

    function readQuestionAnswer(questionID) {
        if (!isPositiveInteger(questionID)) {
            return null;
        }

        const currentRaw = localStorage.getItem(questionAnswerPrefix + questionID);
        if (currentRaw !== null) {
            return parseQuestionRecord(currentRaw, "", false);
        }

        const legacyRaw = localStorage.getItem(legacyAnswerPrefix + questionID);
        if (legacyRaw === null) {
            return null;
        }

        const record = parseQuestionRecord(
            legacyRaw,
            localStorage.getItem(legacyUpdatedAtPrefix + questionID),
            localStorage.getItem(legacySubmittedPrefix + questionID)
        );

        // Миграция выполняется при чтении, но старый ключ не удаляется до явного сохранения.
        if (record) {
            try {
                localStorage.setItem(questionAnswerPrefix + questionID, JSON.stringify(record));
            } catch (error) {
                // Даже если запись недоступна, старый ответ остаётся доступен в текущем чтении.
            }
        }

        return record;
    }

    function saveQuestionAnswer(questionID, answer, submitted) {
        if (!isPositiveInteger(questionID)) {
            throw new Error("invalid question id");
        }

        const record = createQuestionRecord(answer, new Date().toISOString(), submitted);
        localStorage.setItem(questionAnswerPrefix + questionID, JSON.stringify(record));
        localStorage.removeItem(legacyAnswerPrefix + questionID);
        localStorage.removeItem(legacyUpdatedAtPrefix + questionID);
        localStorage.removeItem(legacySubmittedPrefix + questionID);
        return record;
    }

    function clearQuestionAnswer(questionID) {
        if (!isPositiveInteger(questionID)) {
            return;
        }

        localStorage.removeItem(questionAnswerPrefix + questionID);
        localStorage.removeItem(legacyAnswerPrefix + questionID);
        localStorage.removeItem(legacyUpdatedAtPrefix + questionID);
        localStorage.removeItem(legacySubmittedPrefix + questionID);
    }

    function isLessonRead(lessonID) {
        if (!isPositiveInteger(lessonID)) {
            return false;
        }
        return normalizeBoolean(localStorage.getItem(lessonReadPrefix + lessonID));
    }

    function markLessonRead(lessonID) {
        if (!isPositiveInteger(lessonID)) {
            throw new Error("invalid lesson id");
        }
        localStorage.setItem(lessonReadPrefix + lessonID, "true");
    }

    function clearLessonRead(lessonID) {
        if (isPositiveInteger(lessonID)) {
            localStorage.removeItem(lessonReadPrefix + lessonID);
        }
    }

    function normalizeChecks(value) {
        const checks = value && typeof value === "object" && !Array.isArray(value) ? value : {};
        return {
            runs: normalizeBoolean(checks.runs),
            matches_requirements: normalizeBoolean(checks.matches_requirements),
            understands_code: normalizeBoolean(checks.understands_code),
            checked_errors: normalizeBoolean(checks.checked_errors),
            checked_success_and_error: normalizeBoolean(checks.checked_success_and_error),
            done_independently: normalizeBoolean(checks.done_independently),
        };
    }

    function normalizePracticeRecord(value) {
        if (!value || typeof value !== "object" || Array.isArray(value)) {
            return null;
        }

        return {
            task_id: Number(value.task_id) || 0,
            lesson_id: Number(value.lesson_id) || 0,
            solution: normalizeString(value.solution),
            explanation: normalizeString(value.explanation),
            execution_result: normalizeString(value.execution_result),
            difficulties: normalizeString(value.difficulties),
            checks: normalizeChecks(value.checks),
            self_assessment: normalizeString(value.self_assessment),
            repository_url: normalizeString(value.repository_url),
            branch: normalizeString(value.branch),
            commit_hash: normalizeString(value.commit_hash),
            updated_at: normalizeString(value.updated_at),
            submitted: normalizeBoolean(value.submitted),
        };
    }

    function readPracticeSubmission(taskID) {
        if (!isPositiveInteger(taskID)) {
            return null;
        }

        const rawValue = localStorage.getItem(practiceSubmissionPrefix + taskID);
        if (rawValue === null) {
            return null;
        }

        try {
            return normalizePracticeRecord(JSON.parse(rawValue));
        } catch (error) {
            return null;
        }
    }

    function savePracticeSubmission(value) {
        const record = normalizePracticeRecord(value);
        if (!record || !isPositiveInteger(record.task_id) || !isPositiveInteger(record.lesson_id)) {
            throw new Error("invalid practice submission");
        }

        record.updated_at = new Date().toISOString();
        localStorage.setItem(practiceSubmissionPrefix + record.task_id, JSON.stringify(record));
        return record;
    }

    function clearPracticeSubmission(taskID) {
        if (isPositiveInteger(taskID)) {
            localStorage.removeItem(practiceSubmissionPrefix + taskID);
        }
    }

    function submittedQuestionCount(questionIDs) {
        if (!Array.isArray(questionIDs)) {
            return 0;
        }

        let count = 0;
        for (const questionID of questionIDs) {
            const record = readQuestionAnswer(Number(questionID));
            if (record && record.submitted && record.answer.trim() !== "") {
                count += 1;
            }
        }
        return count;
    }

    function allQuestionsSubmitted(questionIDs) {
        return Array.isArray(questionIDs)
            && questionIDs.length > 0
            && submittedQuestionCount(questionIDs) === questionIDs.length;
    }

    window.MentorForgeStorage = Object.freeze({
        readQuestionAnswer: readQuestionAnswer,
        saveQuestionAnswer: saveQuestionAnswer,
        clearQuestionAnswer: clearQuestionAnswer,
        isLessonRead: isLessonRead,
        markLessonRead: markLessonRead,
        clearLessonRead: clearLessonRead,
        readPracticeSubmission: readPracticeSubmission,
        savePracticeSubmission: savePracticeSubmission,
        clearPracticeSubmission: clearPracticeSubmission,
        submittedQuestionCount: submittedQuestionCount,
        allQuestionsSubmitted: allQuestionsSubmitted,
    });
})();
