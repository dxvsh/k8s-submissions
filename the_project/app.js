const API_URL = `${window.location.protocol}//${window.location.hostname}:8081/api/todos`;

const todoForm = document.querySelector("#todo-form");
const todoInput = document.querySelector("#todo-input");
const todoList = document.querySelector("#todo-list");
const sendButton = document.querySelector("#send-button");
const formError = document.querySelector("#form-error");

function showError(message) {
	formError.textContent = message;
}

function renderTodos(todos) {
	todoList.replaceChildren();

	if (todos.length === 0) {
		const emptyState = document.createElement("li");
		emptyState.className = "empty-state";
		emptyState.textContent = "No todos yet. Add one above!";
		todoList.appendChild(emptyState);
		return;
	}

	for (const todo of todos) {
		const item = document.createElement("li");
		item.className = "todo-item";
		item.textContent = todo.text;
		todoList.appendChild(item);
	}
}

async function loadTodos() {
	try {
		const response = await fetch(API_URL);
		if (!response.ok) {
			throw new Error(`Backend returned status ${response.status}`);
		}

		const todos = await response.json();
		renderTodos(todos);
		showError("");
	} catch (error) {
		console.error("Failed to load todos:", error);
		showError("Could not load todos. Make sure the backend is running on port 8081.");
	}
}

async function createTodo(text) {
	const response = await fetch(API_URL, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ text }),
	});

	if (!response.ok) {
		const responseBody = await response.json().catch(() => ({}));
		throw new Error(responseBody.error || `Backend returned status ${response.status}`);
	}

	return response.json();
}

todoForm.addEventListener("submit", async (event) => {
	event.preventDefault();
	showError("");

	const text = todoInput.value.trim();
	if (text.length === 0) {
		showError("Please enter a todo.");
		return;
	}
	if ([...text].length > 140) {
		showError("A todo cannot exceed 140 characters.");
		return;
	}

	sendButton.disabled = true;
	try {
		await createTodo(text);
		todoInput.value = "";
		await loadTodos();
		todoInput.focus();
	} catch (error) {
		console.error("Failed to create todo:", error);
		showError(error.message);
	} finally {
		sendButton.disabled = false;
	}
});

loadTodos();
