describe('login', () => {
  beforeEach(() => {
      cy.visit('/login/');
  });
  
  it('should require cookies to be accepted', () => {
    cy.get('#username').click().type('admin');
    cy.get('#password').click().type('admin');
    cy.get('#login_btn').click();
    cy.get('#notif').should('be.visible').should('contain.text', 'cookie');
  });

  describe('with accepted cookies', () => {
    beforeEach(() => {
      cy.get('gdpr-popup').shadow().find('[data-choice=necessary]').click();
    });

    it('should redirect', () => {
      cy.get('#username').click().type('admin');
      cy.get('#password').click().type('admin');
      cy.get('#login_btn').click();
      
      cy.url().should('eq', Cypress.config().baseUrl + '/search/')
    });
  });
});
