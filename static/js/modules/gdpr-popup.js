import { new_template, hide } from "./util.js";
import { } from "./popup-container.js"; // for import order, and so we dont have to import this script in html

const key = 'gdpr-consent';

export let consent = localStorage.getItem(key);

const make_body = new_template(`
<p>Deze site gebruikt cookies voor functionaliteiten zoals inloggen. Accepteer je de cookies? </p>
<div class="flex-row-center g20" style="flex-wrap: wrap; padding-top: 20px;">
	<div class="flex-row-center g20" style="flex-wrap: wrap; align-items: center;">
		<button data-choice="yes">Ja, eet mijn cookies a.u.b.</button>
		<button data-choice="necessary">Alleen functionele cookies.</button>
		<button data-choice="later">Boeie, vraag later.</button>
	</div>
	<img src="/img/cookie.png" style="max-width: fit-content; max-height: fit-content;">
</div>
<link rel="stylesheet" href="/css/reset.css">
<link rel="stylesheet" href="/css/style.css">
<style>
:host {
	background: white;
	border: 5px solid black;
	padding: 20px;
	max-width: 900px;
	filter: drop-shadow(0 0 5px black);
}
</style>
`)

function cookies_allowed() {
	return consent === "yes" || consent === "necessary";
}

export class GDPRPopup extends HTMLElement {
	static { customElements.define("gdpr-popup", GDPRPopup); }

	constructor() {
		super();

		this.attachShadow({mode: 'open'});
		this.shadowRoot.append(make_body());

		const should_hide = cookies_allowed();
		hide(this, should_hide)

		const handler = e => {
			const choice = e.target.dataset["choice"];
			localStorage.setItem(key, choice);
			consent = choice;
			if (cookies_allowed()) {
				document.cookie = `gdpr-consent=${choice}; path=/; expires=Fri, 31 Dec 9999 23:59:59 GMT`;
			}
			hide(this, true);
		}

		const btns = this.shadowRoot.querySelectorAll("button[data-choice]");
		for (let i = 0; i < btns.length; i++) {
			btns[i].addEventListener("click", handler);
		}
	}
}