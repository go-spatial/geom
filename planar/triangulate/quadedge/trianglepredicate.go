/*
Copyright (c) 2016 Vivid Solutions.

All rights reserved. This program and the accompanying materials
are made available under the terms of the Eclipse Public License v1.0
and Eclipse Distribution License v. 1.0 which accompanies this distribution.
The Eclipse Public License is available at http://www.eclipse.org/legal/epl-v10.html
and the Eclipse Distribution License is available at

http://www.eclipse.org/org/documents/edl-v10.php.
*/

package quadedge

import (
	"github.com/go-spatial/geom"
)

/*
Algorithms for computing values and predicates associated with triangles.
For some algorithms extended-precision implementations are provided, which are
more robust (i.e. they produce correct answers in more cases). Also, some more
robust formulations of some algorithms are provided, which utilize
normalization to the origin.

Author Martin Davis
Ported to Go by Jason R. Surratt

The empty struct gives us a "static" TrianglePredicate namespace.
*/
type trianglePredicate struct{}

var TrianglePredicate trianglePredicate

/**
   * Tests if a point is inside the circle defined by
   * the triangle with vertices a, b, c (oriented counter-clockwise).
   * This test uses simple
   * double-precision arithmetic, and thus may not be robust.
   *
   * @param a a vertex of the triangle
   * @param b a vertex of the triangle
   * @param c a vertex of the triangle
   * @param p the point to test
   * @return true if this point is inside the circle defined by the points a, b, c
  public static boolean isInCircleNonRobust(
      Coordinate a, Coordinate b, Coordinate c,
      Coordinate p) {
    boolean isInCircle =
              (a.x * a.x + a.y * a.y) * triArea(b, c, p)
            - (b.x * b.x + b.y * b.y) * triArea(a, c, p)
            + (c.x * c.x + c.y * c.y) * triArea(a, b, p)
            - (p.x * p.x + p.y * p.y) * triArea(a, b, c)
            > 0;
    return isInCircle;
  }
*/

/*
Tests if a point is inside the circle defined by
the triangle with vertices a, b, c (oriented counter-clockwise).
This test uses simple
double-precision arithmetic, and thus is not 100% robust.
However, by using normalization to the origin
it provides improved robustness and increased performance.

Based on code by J.R.Shewchuk.

a - a vertex of the triangle
b - a vertex of the triangle
c - a vertex of the triangle
p - the point to test
Returns true if this point is inside the circle defined by the points a, b, c
*/
func (_ trianglePredicate) IsInCircleNormalized(a geom.Pointer, b geom.Pointer, c geom.Pointer, p geom.Pointer) bool {
	adx := a.XY()[0] - p.XY()[0]
	ady := a.XY()[1] - p.XY()[1]
	bdx := b.XY()[0] - p.XY()[0]
	bdy := b.XY()[1] - p.XY()[1]
	cdx := c.XY()[0] - p.XY()[0]
	cdy := c.XY()[1] - p.XY()[1]

	abdet := adx*bdy - bdx*ady
	bcdet := bdx*cdy - cdx*bdy
	cadet := cdx*ady - adx*cdy
	alift := adx*adx + ady*ady
	blift := bdx*bdx + bdy*bdy
	clift := cdx*cdx + cdy*cdy

	disc := alift*bcdet + blift*cadet + clift*abdet
	return disc > 0
}

/**
   * Computes twice the area of the oriented triangle (a, b, c), i.e., the area is positive if the
   * triangle is oriented counterclockwise.
   *
   * @param a a vertex of the triangle
   * @param b a vertex of the triangle
   * @param c a vertex of the triangle
  private static double triArea(Coordinate a, Coordinate b, Coordinate c) {
      return (b.x - a.x) * (c.y - a.y)
           - (b.y - a.y) * (c.x - a.x);
  }
*/

/*
Tests if a point is inside the circle defined by
the triangle with vertices a, b, c (oriented counter-clockwise).
This method uses more robust computation.

a - a vertex of the triangle
b - a vertex of the triangle
c - a vertex of the triangle
p - the point to test
Returns true if this point is inside the circle defined by the points a, b, c
*/
func (tp trianglePredicate) IsInCircleRobust(a geom.Pointer, b geom.Pointer, c geom.Pointer, p geom.Pointer) bool {
	//checkRobustInCircle(a, b, c, p);
	//    return isInCircleNonRobust(a, b, c, p);
	return tp.IsInCircleNormalized(a, b, c, p)
}

/**
   * Tests if a point is inside the circle defined by
   * the triangle with vertices a, b, c (oriented counter-clockwise).
   * The computation uses {@link DD} arithmetic for robustness.
   *
   * @param a a vertex of the triangle
   * @param b a vertex of the triangle
   * @param c a vertex of the triangle
   * @param p the point to test
   * @return true if this point is inside the circle defined by the points a, b, c
  public static boolean isInCircleDDSlow(
      Coordinate a, Coordinate b, Coordinate c,
      Coordinate p) {
    DD px = DD.valueOf(p.x);
    DD py = DD.valueOf(p.y);
    DD ax = DD.valueOf(a.x);
    DD ay = DD.valueOf(a.y);
    DD bx = DD.valueOf(b.x);
    DD by = DD.valueOf(b.y);
    DD cx = DD.valueOf(c.x);
    DD cy = DD.valueOf(c.y);

    DD aTerm = (ax.multiply(ax).add(ay.multiply(ay)))
        .multiply(triAreaDDSlow(bx, by, cx, cy, px, py));
    DD bTerm = (bx.multiply(bx).add(by.multiply(by)))
        .multiply(triAreaDDSlow(ax, ay, cx, cy, px, py));
    DD cTerm = (cx.multiply(cx).add(cy.multiply(cy)))
        .multiply(triAreaDDSlow(ax, ay, bx, by, px, py));
    DD pTerm = (px.multiply(px).add(py.multiply(py)))
        .multiply(triAreaDDSlow(ax, ay, bx, by, cx, cy));

    DD sum = aTerm.subtract(bTerm).add(cTerm).subtract(pTerm);
    boolean isInCircle = sum.doubleValue() > 0;

    return isInCircle;
  }
*/

/**
   * Computes twice the area of the oriented triangle (a, b, c), i.e., the area
   * is positive if the triangle is oriented counterclockwise.
   * The computation uses {@link DD} arithmetic for robustness.
   *
   * @param ax the x ordinate of a vertex of the triangle
   * @param ay the y ordinate of a vertex of the triangle
   * @param bx the x ordinate of a vertex of the triangle
   * @param by the y ordinate of a vertex of the triangle
   * @param cx the x ordinate of a vertex of the triangle
   * @param cy the y ordinate of a vertex of the triangle
  public static DD triAreaDDSlow(DD ax, DD ay,
      DD bx, DD by, DD cx, DD cy) {
    return (bx.subtract(ax).multiply(cy.subtract(ay)).subtract(by.subtract(ay)
        .multiply(cx.subtract(ax))));
  }

  public static boolean isInCircleDDFast(
      Coordinate a, Coordinate b, Coordinate c,
      Coordinate p) {
    DD aTerm = (DD.sqr(a.x).selfAdd(DD.sqr(a.y)))
        .selfMultiply(triAreaDDFast(b, c, p));
    DD bTerm = (DD.sqr(b.x).selfAdd(DD.sqr(b.y)))
        .selfMultiply(triAreaDDFast(a, c, p));
    DD cTerm = (DD.sqr(c.x).selfAdd(DD.sqr(c.y)))
        .selfMultiply(triAreaDDFast(a, b, p));
    DD pTerm = (DD.sqr(p.x).selfAdd(DD.sqr(p.y)))
        .selfMultiply(triAreaDDFast(a, b, c));

    DD sum = aTerm.selfSubtract(bTerm).selfAdd(cTerm).selfSubtract(pTerm);
    boolean isInCircle = sum.doubleValue() > 0;

    return isInCircle;
  }

  public static DD triAreaDDFast(
      Coordinate a, Coordinate b, Coordinate c) {

    DD t1 = DD.valueOf(b.x).selfSubtract(a.x)
          .selfMultiply(
              DD.valueOf(c.y).selfSubtract(a.y));

    DD t2 = DD.valueOf(b.y).selfSubtract(a.y)
          .selfMultiply(
              DD.valueOf(c.x).selfSubtract(a.x));

    return t1.selfSubtract(t2);
  }

  public static boolean isInCircleDDNormalized(
      Coordinate a, Coordinate b, Coordinate c,
      Coordinate p) {
    DD adx = DD.valueOf(a.x).selfSubtract(p.x);
    DD ady = DD.valueOf(a.y).selfSubtract(p.y);
    DD bdx = DD.valueOf(b.x).selfSubtract(p.x);
    DD bdy = DD.valueOf(b.y).selfSubtract(p.y);
    DD cdx = DD.valueOf(c.x).selfSubtract(p.x);
    DD cdy = DD.valueOf(c.y).selfSubtract(p.y);

    DD abdet = adx.multiply(bdy).selfSubtract(bdx.multiply(ady));
    DD bcdet = bdx.multiply(cdy).selfSubtract(cdx.multiply(bdy));
    DD cadet = cdx.multiply(ady).selfSubtract(adx.multiply(cdy));
    DD alift = adx.multiply(adx).selfAdd(ady.multiply(ady));
    DD blift = bdx.multiply(bdx).selfAdd(bdy.multiply(bdy));
    DD clift = cdx.multiply(cdx).selfAdd(cdy.multiply(cdy));

    DD sum = alift.selfMultiply(bcdet)
    .selfAdd(blift.selfMultiply(cadet))
    .selfAdd(clift.selfMultiply(abdet));

    boolean isInCircle = sum.doubleValue() > 0;

    return isInCircle;
  }
*/

/**
   * Computes the inCircle test using distance from the circumcentre.
   * Uses standard double-precision arithmetic.
   * <p>
   * In general this doesn't
   * appear to be any more robust than the standard calculation. However, there
   * is at least one case where the test point is far enough from the
   * circumcircle that this test gives the correct answer.
   * <pre>
   * LINESTRING
   * (1507029.9878 518325.7547, 1507022.1120341457 518332.8225183258,
   * 1507029.9833 518325.7458, 1507029.9896965567 518325.744909031)
   * </pre>
   *
   * @param a a vertex of the triangle
   * @param b a vertex of the triangle
   * @param c a vertex of the triangle
   * @param p the point to test
   * @return true if this point is inside the circle defined by the points a, b, c
  public static boolean isInCircleCC(Coordinate a, Coordinate b, Coordinate c,
      Coordinate p) {
    Coordinate cc = Triangle.circumcentre(a, b, c);
    double ccRadius = a.distance(cc);
    double pRadiusDiff = p.distance(cc) - ccRadius;
    return pRadiusDiff <= 0;
  }
*/

/**
   * Checks if the computed value for isInCircle is correct, using
   * double-double precision arithmetic.
   *
   * @param a a vertex of the triangle
   * @param b a vertex of the triangle
   * @param c a vertex of the triangle
   * @param p the point to test
private static void checkRobustInCircle(Coordinate a, Coordinate b, Coordinate c,
    Coordinate p)
{
  boolean nonRobustInCircle = isInCircleNonRobust(a, b, c, p);
  boolean isInCircleDD = TrianglePredicate.isInCircleDDSlow(a, b, c, p);
  boolean isInCircleCC = TrianglePredicate.isInCircleCC(a, b, c, p);

  Coordinate circumCentre = Triangle.circumcentre(a, b, c);
  System.out.println("p radius diff a = "
      + Math.abs(p.distance(circumCentre) - a.distance(circumCentre))
      / a.distance(circumCentre));

  if (nonRobustInCircle != isInCircleDD || nonRobustInCircle != isInCircleCC) {
    System.out.println("inCircle robustness failure (double result = "
        + nonRobustInCircle
        + ", DD result = " + isInCircleDD
        + ", CC result = " + isInCircleCC + ")");
    System.out.println(WKTWriter.toLineString(new CoordinateArraySequence(
        new Coordinate[] { a, b, c, p })));
    System.out.println("Circumcentre = " + WKTWriter.toPoint(circumCentre)
        + " radius = " + a.distance(circumCentre));
    System.out.println("p radius diff a = "
        + Math.abs(p.distance(circumCentre)/a.distance(circumCentre) - 1));
    System.out.println("p radius diff b = "
        + Math.abs(p.distance(circumCentre)/b.distance(circumCentre) - 1));
    System.out.println("p radius diff c = "
        + Math.abs(p.distance(circumCentre)/c.distance(circumCentre) - 1));
    System.out.println();
  }
}


}
*/
