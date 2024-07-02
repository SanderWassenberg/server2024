const selector = 'gdpr-popup';
const cookie_name = 'gdpr-consent';

describe('gdpr popup', () => {
  beforeEach(() => {
    cy.visit('/');
  });

  it('should not have set cookie', () => {
    cy.getCookie(cookie_name).should("not.exist");
  });
  it('should be visible', () => {
    cy.get(selector).should('be.visible');
  });

  for (const choice of ["yes", "necessary", "later"]) {
    describe(`when pressing "${choice}"`, () => {
      beforeEach(() => {
        cy.get(selector).shadow().find(`[data-choice=${choice}]`).click();
      });

      it('should be visible', () => {
        cy.get(selector).should('not.be.visible');
      });

      it('should set cookie as expected', () => {
        if (choice == "later") {
          cy.getCookie(cookie_name).should("not.exist");
        } else {
          cy.getCookie(cookie_name).should("have.property", "value", choice);
        }
      });

      it('should be remember choice', () => {
        cy.reload();
        if (choice == "later") {
          cy.get(selector).should("be.visible");
        } else {
          cy.get(selector).should("not.be.visible");
        }
      });
    });
  }
});