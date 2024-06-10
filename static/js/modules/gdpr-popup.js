import { new_template, hide } from "./util.js"

const key = 'gdpr-consent';

export let consent = localStorage.getItem(key);

const make_body = new_template(`
<p>Hee hoi, wij willen graag al jouw cookies opeten, vind je dat erg? Alsje alsje alsjeblieeeeft? Om nom nom nom, cookies!</p>
<div class="flex-row-center" style="flex-wrap: wrap; gap: 20px; padding-top: 20px;">
	<div class="flex-row-center" style="flex-wrap: wrap; gap: 20px; align-items: center;">
		<button data-choice="yes">Ja, eet mijn cookies a.u.b.</button>
		<button data-choice="no">Nee, blijf van mijn cookies af!</button>
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


export class GDPRPopup extends HTMLElement {
	static { customElements.define("gdpr-popup", GDPRPopup); }

	constructor() {
		super();

		this.attachShadow({mode: 'open'});
		this.shadowRoot.append(make_body());

		const should_hide = consent === "yes" || consent === "no";
		hide(this, should_hide)

		const handler = e => {
			const choice = e.target.dataset["choice"];
			localStorage.setItem(key, choice)
			consent = choice;
			hide(this, true);
		}

		const btns = this.shadowRoot.querySelectorAll("button[data-choice]");
		for (let i = 0; i < btns.length; i++) {
			btns[i].addEventListener("click", handler);
		}
	}
}