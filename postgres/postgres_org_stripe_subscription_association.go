package postgres

import (
	"context"
	"vector-ai/model"
)

func (pgx Pgx) CreateOrgStripeSubscriptionAssociation(stripeId string, orgId string, active bool) (model.OrgStripeSubscriptionAssociation, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), "INSERT INTO org_stripe_subscription_associations (stripe_subscription_id, org_id, active) VALUES ($1, $2, $3)", stripeId, orgId, active)
	if err != nil || commandTag.RowsAffected() != 1 {
		var orgStripe model.OrgStripeSubscriptionAssociation
		return orgStripe, err
	}
	return pgx.GetOrgStripeSubscriptionAssociationByOrgId(orgId)
}

func (pgx Pgx) GetOrgStripeSubscriptionAssociationByOrgId(orgId string) (model.OrgStripeSubscriptionAssociation, error) {
	var orgStripe model.OrgStripeSubscriptionAssociation

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT stripe_subscription_id, org_id, active FROM org_stripe_subscription_associations WHERE org_id=$1", orgId).Scan(&orgStripe.StripeSubscriptionID, &orgStripe.OrgID, &orgStripe.Active); err != nil {
		return orgStripe, err
	}
	return orgStripe, nil
}

func (pgx Pgx) GetOrgStripeSubscriptionAssociationByStripeId(stripeId string) (model.OrgStripeSubscriptionAssociation, error) {
	var orgStripe model.OrgStripeSubscriptionAssociation

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT stripe_subscription_id, org_id, active FROM org_stripe_subscription_associations WHERE stripe_subscription_id=$1", stripeId).Scan(&orgStripe.StripeSubscriptionID, &orgStripe.OrgID, &orgStripe.Active); err != nil {
		return orgStripe, err
	}
	return orgStripe, nil
}

func (pgx Pgx) UpdateOrgStripeSubscriptionAssociation(active bool, stripeId string) (model.OrgStripeSubscriptionAssociation, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE org_stripe_subscription_associations SET active=$1 WHERE stripe_subscription_id=$2", active, stripeId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var orgStripe model.OrgStripeSubscriptionAssociation
		return orgStripe, err
	}

	return pgx.GetOrgStripeSubscriptionAssociationByStripeId(stripeId)
}

func (pgx Pgx) DeleteOrgStripeSubscriptionAssociation(orgId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM org_stripe_subscription_associations WHERE org_id=$1", orgId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
